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

package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// organizatonRoleMembershipResource is the resource implementation.
type organizationRoleMembershipResource struct {
	baseResource
}

type organizationRoleMembershipModelResource struct {
	ID             types.String `tfsdk:"id"`
	RoleId         types.String `tfsdk:"role_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	UserName       types.String `tfsdk:"user_name"`
	GroupName      types.String `tfsdk:"group_name"`
}

// NewOrganizationRoleMembershipResource is a helper function to simplify the provider implementation.
func NewOrganizationRoleMembershipResource() resource.Resource {
	return &organizationRoleMembershipResource{}
}

// Metadata returns the resource type name.
func (r *organizationRoleMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_role_membership"
}

// Schema defines the schema for the resource.
func (r *organizationRoleMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"role_id": schema.StringAttribute{
				Required:    true,
				Description: "The role ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_id": schema.StringAttribute{
				Required:    true,
				Description: "The organization ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_name": schema.StringAttribute{
				Optional:    true,
				Description: "The username of the user (mutually exclusive with group_name)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_name": schema.StringAttribute{
				Optional:    true,
				Description: "The group name of the group (mutually exclusive with user_name)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *organizationRoleMembershipResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("user_name"),
			path.MatchRoot("group_name"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *organizationRoleMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data organizationRoleMembershipModelResource

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to create organization role membership
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.auth,
	)

	// Determine the member type, which can be any of group or user.
	// The resource validator makes sure that exactly one of these is configured.
	var memberType, memberName string
	if !data.GroupName.IsNull() {
		memberType = "group"
		memberName = data.GroupName.ValueString()
	} else {
		memberType = "user"
		memberName = data.UserName.ValueString()
	}

	apiRequest := r.client.RoleMembershipsAPI.GrantRoleMembershipApplicationOrOrganization(ctx, "organization", data.OrganizationId.ValueString(), data.RoleId.ValueString(), memberType, memberName)
	apiResponse, err := r.client.RoleMembershipsAPI.GrantRoleMembershipApplicationOrOrganizationExecute(apiRequest)

	// Call API
	if err != nil {
		error_body, _ := io.ReadAll(apiResponse.Body)
		resp.Diagnostics.AddError(
			"Error creating organization role membership",
			"Could not create organization role membership, unexpected error: "+apiResponse.Status+": "+string(error_body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values.
	// Because the organization role membership does not have an ID of its own, we create a synthetic one based on the provided attributes.
	data.ID = types.StringValue(fmt.Sprintf("%s_%s_%s_%s", data.OrganizationId.ValueString(), data.RoleId.ValueString(), memberType, memberName))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *organizationRoleMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data organizationRoleMembershipModelResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.auth,
	)

	// Get refreshed organization role membership from IQ
	apiRequest := r.client.RoleMembershipsAPI.GetRoleMembershipsApplicationOrOrganization(ctx, "organization", data.OrganizationId.ValueString())
	roleMemberships, apiResponse, err := r.client.RoleMembershipsAPI.GetRoleMembershipsApplicationOrOrganizationExecute(apiRequest)

	// Check if we received a list of role mappings.
	if err != nil {
		if apiResponse.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error Reading IQ organization role membership",
				"Could not read organization role membership with ID "+data.ID.ValueString()+": "+err.Error(),
			)
		}
		return
	}

	// Determine the member type, which can be any of group or user.
	// The resource validator makes sure that exactly one of these is configured.
	var memberType, memberName string
	if !data.GroupName.IsNull() {
		memberType = "GROUP"
		memberName = data.GroupName.ValueString()
	} else {
		memberType = "USER"
		memberName = data.UserName.ValueString()
	}

	// Check for organization role membership existence.
	var membershipFound bool
	for _, roleMembership := range roleMemberships.MemberMappings {
		if *roleMembership.RoleId == data.RoleId.ValueString() {
			for _, member := range roleMembership.Members {
				if *member.Type == memberType && *member.UserOrGroupName == memberName && *member.OwnerType == "ORGANIZATION" && *member.OwnerId == data.OrganizationId.ValueString() {
					membershipFound = true
					break
				}
			}
		}
	}

	if !membershipFound {
		resp.State.RemoveResource(ctx)
		return
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *organizationRoleMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data organizationRoleMembershipModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make Delete API Call
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.auth,
	)

	// Determine the member type, which can be any of group or user.
	// The resource validator makes sure that exactly one of these is configured.
	var memberType, memberName string
	if !data.GroupName.IsNull() {
		memberType = "group"
		memberName = data.GroupName.ValueString()
	} else {
		memberType = "user"
		memberName = data.UserName.ValueString()
	}

	apiRequest := r.client.RoleMembershipsAPI.RevokeRoleMembershipApplicationOrOrganization(ctx, "organization", data.OrganizationId.ValueString(), data.RoleId.ValueString(), memberType, memberName)
	apiResponse, err := r.client.RoleMembershipsAPI.RevokeRoleMembershipApplicationOrOrganizationExecute(apiRequest)
	if err != nil {
		error_body, _ := io.ReadAll(apiResponse.Body)
		resp.Diagnostics.AddError(
			"Error deleting organization role membership",
			"Could not delete organization role membership, unexpected error: "+apiResponse.Status+": "+string(error_body),
		)
		return
	}
}
