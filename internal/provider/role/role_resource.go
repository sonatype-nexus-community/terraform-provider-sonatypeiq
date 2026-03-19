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
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// roleResource is the resource implementation.
type roleResource struct {
	common.BaseResource
}

// NewRoleResource is a helper function to simplify the provider implementation.
func NewRoleResource() resource.Resource {
	return &roleResource{}
}

// Metadata returns the resource type name.
func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Schema defines the schema for the resource.
func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this to manage Custom Roles",
		Attributes: map[string]schema.Attribute{
			"id":          sharedrschema.ResourceComputedString("Internal Role ID for Terraform State"),
			"name":        sharedrschema.ResourceRequiredString("Role Name"),
			"description": sharedrschema.ResourceRequiredString("Role Description"),
			"built_in":    sharedrschema.ResourceComputedBool("Whether this is a built-in Role in Sonatype IQ"),
			"permissions": sharedrschema.ResourceRequiredSingleNestedAttribute(
				"Permissions for this Role",
				map[string]schema.Attribute{
					"admin": sharedrschema.ResourceRequiredSingleNestedAttribute(
						"Administrator Permmissions",
						map[string]schema.Attribute{
							"access_audit_log": sharedrschema.ResourceOptionalBoolWithDefault(
								"Access to Audit Logs",
								false,
							),
							"view_roles": sharedrschema.ResourceOptionalBoolWithDefault(
								"View all Roles",
								false,
							),
						},
					),
					"iq": sharedrschema.ResourceRequiredSingleNestedAttribute(
						"Sonatype IQ Permmissions",
						map[string]schema.Attribute{
							"add_applications": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can add Applications",
								false,
							),
							"claim_components": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can Claim Components",
								false,
							),
							"edit_access_control": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can edit Access Control",
								false,
							),
							"edit_iq_elements": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can edit IQ Elements",
								false,
							),
							"edit_proprietary_components": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can edit Proprietary Components",
								false,
							),
							"evaluate_applications": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can Evaluate Applications",
								false,
							),
							"evaluate_individual_components": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can Evaluate Individual Components",
								false,
							),
							"manage_automatic_application_creation": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can manage Automatic Application creation",
								false,
							),
							"manage_automatic_scm_configuration": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can manage Automatic SCM Configuration",
								false,
							),
							"view_iq_elements": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can view IQ Elements",
								false,
							),
						},
					),
					"remediation": sharedrschema.ResourceRequiredSingleNestedAttribute(
						"Remediation Permmissions",
						map[string]schema.Attribute{
							"change_licenses": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can change Licenses",
								false,
							),
							"change_security_vulnerabilities": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can change Security Vulnerabilities",
								false,
							),
							"create_pull_requests": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can create Pull Requests",
								false,
							),
							"review_legal_obligations": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can review Legal Obligations",
								false,
							),
							"waive_policy_violations": sharedrschema.ResourceOptionalBoolWithDefault(
								"Can Waive Policy Violations",
								false,
							),
						},
					),
				},
			),
			"last_updated": sharedrschema.ResourceLastUpdated(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.RoleModelResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := r.Client.RolesAPI.AddRole(r.AuthContext(ctx)).ApiRoleDTO(*plan.MapToApi(true)).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating Role",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.HandleAPIError(
			"Creation of Role was not successful",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// Update State based on Response
	plan.MapFromApi(apiResponse)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.RoleModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := r.Client.RolesAPI.GetRoleById(r.AuthContext(ctx), state.ID.ValueString()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error reading Role",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.HandleAPIError(
			"Reading Role by ID was not successful",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// Update State based on Response
	state.MapFromApi(apiResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// // Update updates the resource and sets the updated Terraform state on success.
func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.RoleModelResource
	var state model.RoleModelResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	plan.ID = state.ID
	apiResponse, httpResponse, err := r.Client.RolesAPI.UpdateRole(r.AuthContext(ctx), state.ID.ValueString()).ApiRoleDTO(*plan.MapToApi(true)).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error updating Role",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.HandleAPIError(
			"Updating Role was not successful",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// Map response to State
	state.MapFromApi(apiResponse)

	// Update State
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state model.RoleModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	httpResponse, err := r.Client.RolesAPI.DeleteRole(r.AuthContext(ctx), state.ID.ValueString()).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			fmt.Sprintf(common.ERR_ROLE_DID_NOT_EXIST, state.ID.ValueString()),
			fmt.Sprintf("%v", err),
		)
		return
	}
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
