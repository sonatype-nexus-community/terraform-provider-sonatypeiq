/*
 * Copyright (c) 2019-present Sonatype, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package role

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// globalRoleMembershipResource is the resource implementation.
type globalRoleMembershipResource struct {
	common.BaseResource
}

// NewGlobalRoleMembershipResource is a helper function to simplify the provider implementation.
func NewGlobalRoleMembershipResource() resource.Resource {
	return &globalRoleMembershipResource{}
}

// Metadata returns the resource type name.
func (r *globalRoleMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_role_membership"
}

// Schema defines the schema for the resource.
func (r *globalRoleMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this resource to grant a global role, such as System Administrator, to a user or a group. " +
			"Global roles can only be granted globally: use `sonatypeiq_organization_role_membership` or " +
			"`sonatypeiq_application_role_membership` for any other role. " +
			"Requires the _Edit System Configuration and Users_ permission.",
		Attributes: map[string]schema.Attribute{
			"id":           sharedrschema.ResourceComputedString("The ID of this resource."),
			"role_id":      sharedrschema.ResourceRequiredStringWithPlanModifier("The ID of the global role", []planmodifier.String{stringplanmodifier.RequiresReplace()}),
			"user_name":    sharedrschema.ResourceOptionalStringWithPlanModifier("The username of the user (mutually exclusive with group_name)", stringplanmodifier.RequiresReplace()),
			"group_name":   sharedrschema.ResourceOptionalStringWithPlanModifier("The name of the group (mutually exclusive with user_name)", stringplanmodifier.RequiresReplace()),
			"last_updated": sharedrschema.ResourceLastUpdated(),
		},
	}
}

func (r *globalRoleMembershipResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("user_name"),
			path.MatchRoot("group_name"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *globalRoleMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.GlobalRoleMembershipModelResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
		return
	}

	// The resource validator makes sure that exactly one of these is configured.
	memberType, memberName := globalMemberTypeAndName(&plan)

	httpResponse, err := r.Client.RoleMembershipsAPI.GrantRoleMembershipGlobalOrRepositoryContainer(
		r.AuthContext(ctx),
		common.OWNER_TYPE_GLOBAL,
		plan.RoleId.ValueString(),
		memberType,
		memberName,
	).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		if summary, detail, ok := common.RoleMembershipScopeAdvice(err); ok {
			resp.Diagnostics.AddError(summary, detail)
			return
		}
		errors.HandleAPIError("Error creating global role membership", &err, httpResponse, &resp.Diagnostics)
		return
	}

	// A role membership does not have an ID of its own, so create a synthetic one from the provided attributes.
	plan.ID = types.StringValue(fmt.Sprintf("%s,%s,%s", plan.RoleId.ValueString(), memberType, memberName))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *globalRoleMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.GlobalRoleMembershipModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	memberType, memberName := globalMemberTypeAndName(&state)

	apiResponse, httpResponse, err := r.Client.RoleMembershipsAPI.GetRoleMembershipsGlobalOrRepositoryContainer(
		r.AuthContext(ctx),
		common.OWNER_TYPE_GLOBAL,
	).Execute()

	if err != nil {
		resp.State.RemoveResource(ctx)
		errors.HandleAPIWarning("Global Role Mappings could not be read", &err, httpResponse, &resp.Diagnostics)
		return
	}

	if !globalMembershipExists(apiResponse, state.RoleId.ValueString(), memberType, memberName) {
		resp.State.RemoveResource(ctx)
		errors.HandleAPIWarning("Global Role Mapping not found", &err, httpResponse, &resp.Diagnostics)
		return
	}

	// During Import - ID will be nil - so set it
	if state.ID.IsNull() {
		state.ID = types.StringValue(fmt.Sprintf("%s,%s,%s", state.RoleId.ValueString(), memberType, memberName))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is never called: every attribute forces a replacement.
func (r *globalRoleMembershipResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Global role memberships cannot be updated in place", "Every attribute forces the membership to be replaced.")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *globalRoleMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.GlobalRoleMembershipModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	memberType, memberName := globalMemberTypeAndName(&state)

	httpResponse, err := r.Client.RoleMembershipsAPI.RevokeRoleMembershipGlobalOrRepositoryContainer(
		r.AuthContext(ctx),
		common.OWNER_TYPE_GLOBAL,
		state.RoleId.ValueString(),
		memberType,
		memberName,
	).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			fmt.Sprintf(common.ERR_FAILED_DELETING_GLOBAL_ROLE_MAPPING, state.ID.ValueString()),
			fmt.Sprintf("%v", err),
		)
	}
}

// ImportState imports a membership. The key is "<role id>,<user|group>,<user or group name>".
func (r *globalRoleMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: <role-internal-id>,[group|user],<username-or-group-name> - Got: %q", req.ID),
		)
		return
	}

	switch strings.ToLower(idParts[1]) {
	case common.MEMBER_TYPE_GROUP:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_name"), idParts[2])...)
	case common.MEMBER_TYPE_USER:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_name"), idParts[2])...)
	default:
		resp.Diagnostics.AddError("Unexpected Import Identifier", fmt.Sprintf("The member type must be 'group' or 'user' - Got: %q", idParts[1]))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_id"), idParts[0])...)
}

// globalMembershipExists looks for the role + member in the global memberships. The endpoint is already scoped to the
// global context, so the member's owner is not part of the match.
func globalMembershipExists(memberships *sonatypeiq.ApiRoleMemberMappingListDTO, roleId, memberType, memberName string) bool {
	for _, mapping := range memberships.MemberMappings {
		if mapping.RoleId == nil || *mapping.RoleId != roleId {
			continue
		}
		for _, member := range mapping.Members {
			if member.Type != nil && member.UserOrGroupName != nil &&
				strings.EqualFold(*member.Type, memberType) && *member.UserOrGroupName == memberName {
				return true
			}
		}
	}
	return false
}

func globalMemberTypeAndName(m *model.GlobalRoleMembershipModelResource) (string, string) {
	if !m.GroupName.IsNull() && m.GroupName.ValueString() != "" {
		m.UserName = types.StringNull()
		return common.MEMBER_TYPE_GROUP, m.GroupName.ValueString()
	}
	m.GroupName = types.StringNull()
	return common.MEMBER_TYPE_USER, m.UserName.ValueString()
}
