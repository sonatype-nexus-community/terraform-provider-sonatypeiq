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

package application

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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// applicationRoleMembershipResource is the resource implementation.
type applicationRoleMembershipResource struct {
	common.BaseResource
}

// NewApplicationRoleMembershipResource is a helper function to simplify the provider implementation.
func NewApplicationRoleMembershipResource() resource.Resource {
	return &applicationRoleMembershipResource{}
}

// Metadata returns the resource type name.
func (r *applicationRoleMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_role_membership"
}

// Schema defines the schema for the resource.
func (r *applicationRoleMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":             sharedrschema.ResourceComputedString("The ID of this resource."),
			"role_id":        sharedrschema.ResourceRequiredString("The role ID"),
			"application_id": sharedrschema.ResourceRequiredString("The application ID"),
			"user_name":      sharedrschema.ResourceOptionalString("The username of the user (mutually exclusive with group_name)"),
			"group_name":     sharedrschema.ResourceOptionalString("The group name of the group (mutually exclusive with user_name)"),
			"last_updated":   sharedrschema.ResourceLastUpdated(),
		},
	}
}

func (r *applicationRoleMembershipResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("user_name"),
			path.MatchRoot("group_name"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *applicationRoleMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan model.ApplicationRoleMembershipModelResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
		return
	}

	// Determine the member type, which can be any of group or user.
	// The resource validator makes sure that exactly one of these is configured.
	var memberType, memberName string = memberTypeAndName(&plan)

	httpResponse, err := r.Client.RoleMembershipsAPI.GrantRoleMembershipApplicationOrOrganization(
		r.AuthContext(ctx),
		common.OWNER_TYPE_APPLICATION,
		plan.ApplicationId.ValueString(),
		plan.RoleId.ValueString(),
		memberType,
		memberName,
	).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		errors.HandleAPIError(
			"Error creating application role membership",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// Because the application role membership does not have an ID of its own, we create a synthetic one based on the provided attributes.
	plan.ID = types.StringValue(fmt.Sprintf("%s,%s,%s,%s", plan.ApplicationId.ValueString(), plan.RoleId.ValueString(), memberType, memberName))

	// Update State
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *applicationRoleMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.ApplicationRoleMembershipModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	// The resource validator makes sure that exactly one of these is configured.
	var memberType, memberName string = memberTypeAndName(&state)

	apiResponse, httpResponse, err := r.Client.RoleMembershipsAPI.GetRoleMembershipsApplicationOrOrganization(
		r.AuthContext(ctx),
		common.OWNER_TYPE_APPLICATION,
		state.ApplicationId.ValueString(),
	).Execute()

	if err != nil {
		resp.State.RemoveResource(ctx)
		errors.HandleAPIWarning(
			"Role Mappings for Application could not be read",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// Iterate all Role Memberships looking for a match
	var membershipFound bool
	for _, roleMembership := range apiResponse.MemberMappings {
		if *roleMembership.RoleId == state.RoleId.ValueString() {
			for _, member := range roleMembership.Members {
				if strings.ToLower(*member.OwnerType) == common.OWNER_TYPE_APPLICATION && *member.OwnerId == state.ApplicationId.ValueString() {
					if strings.ToLower(*member.Type) == memberType && *member.UserOrGroupName == memberName {
						membershipFound = true
						break
					}
				}
			}
		}
	}

	if !membershipFound {
		resp.State.RemoveResource(ctx)
		errors.HandleAPIWarning(
			"Role Mapping not found for Application",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// During Import - ID will be nil - so set it
	if state.ID.IsNull() {
		state.ID = types.StringValue(fmt.Sprintf("%s,%s,%s,%s", state.ApplicationId.ValueString(), state.RoleId.ValueString(), memberType, memberName))
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *applicationRoleMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state model.ApplicationRoleMembershipModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	// The resource validator makes sure that exactly one of these is configured.
	var memberType, memberName string = memberTypeAndName(&state)

	httpResponse, err := r.Client.RoleMembershipsAPI.RevokeRoleMembershipApplicationOrOrganization(
		r.AuthContext(ctx),
		common.OWNER_TYPE_APPLICATION,
		state.ApplicationId.ValueString(),
		state.RoleId.ValueString(),
		memberType,
		memberName,
	).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			fmt.Sprintf(common.ERR_FAILED_DELETING_APPLICATION_ROLE_MAPPING, state.ID.ValueString()),
			fmt.Sprintf("%v", err),
		)
		return
	}
}

// Import
// Key is "%s,%s,%s,%s", plan.ApplicationId.ValueString(), plan.RoleId.ValueString(), memberType, memberName (lower case)
func (r *applicationRoleMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: <application-internal-id>,<role-internal-id>,[group|user],<username-or-group-name> - Got: %q", req.ID),
		)
		return
	}

	switch strings.ToLower(idParts[2]) {
	case common.MEMBER_TYPE_GROUP:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_name"), idParts[3])...)
	case common.MEMBER_TYPE_USER:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_name"), idParts[3])...)
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_id"), idParts[1])...)
}

func memberTypeAndName(state *model.ApplicationRoleMembershipModelResource) (string, string) {
	if !state.GroupName.IsNull() {
		return common.MEMBER_TYPE_GROUP, state.GroupName.ValueString()
	} else {
		return common.MEMBER_TYPE_USER, state.UserName.ValueString()
	}
}
