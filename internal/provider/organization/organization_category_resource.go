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

package organization

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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

// applicationCategoryResource is the resource implementation.
type applicationCategoryResource struct {
	common.BaseResource
}

// NewApplicationCategoryResource is a helper function to simplify the provider implementation.
func NewApplicationCategoryResource() resource.Resource {
	return &applicationCategoryResource{}
}

// Metadata returns the resource type name.
func (r *applicationCategoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_category"
}

// Schema defines the schema for the resource.
func (r *applicationCategoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this resource to manage Application Categories/Tags which can then be applied to Applications.",
		Attributes: map[string]schema.Attribute{
			"id":              sharedrschema.ResourceComputedString("Internal ID of the Application Category"),
			"name":            sharedrschema.ResourceRequiredString("Name of the Application Category"),
			"description":     sharedrschema.ResourceRequiredString("Description of the Application Category"),
			"organization_id": sharedrschema.ResourceRequiredString("Internal ID of the Organization to which this Application Category belongs. Use `ROOT_ORGANIZATION_ID` for the Root Organization."),
			"color": sharedrschema.ResourceRequiredStringEnum(
				"Color of the Application Category",
				model.AllColors()...,
			),
			"last_updated": sharedrschema.ResourceLastUpdated(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *applicationCategoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan model.ApplicationCategoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := r.Client.ApplicationCategoriesAPI.AddTag(
		r.AuthContext(ctx),
		plan.OrganizationId.ValueString(),
	).ApiApplicationCategoryDTO(*plan.MapToApi()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating Application Category",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.HandleAPIError(
			"Creation of Application Category was not successful",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// Map response to State
	plan.MapFromApi(apiResponse)

	// Update State
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *applicationCategoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.ApplicationCategoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := r.Client.ApplicationCategoriesAPI.GetTags(r.AuthContext(ctx), state.OrganizationId.ValueString()).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"Organization with ID did not exist to read Application Categories",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		} else {
			errors.HandleAPIError(
				common.ERR_FAILED_READING_APPLICATION_CATEGORIES_FOR_ORG,
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		}
		return
	}

	if len(apiResponse) == 0 {
		resp.State.RemoveResource(ctx)
		errors.AddValidationDiagnostic(
			&resp.Diagnostics,
			"Application Categories",
			"No Application Categories exist for Organizatio ID",
		)
		return
	}

	// Iterate all Role Memberships looking for a match
	var found = false
	for _, ac := range apiResponse {
		if *ac.Id == state.ID.ValueString() {
			state.MapFromApi(&ac)
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		errors.AddValidationDiagnostic(
			&resp.Diagnostics,
			"Application Category",
			fmt.Sprintf("Application Category with ID %s does not exist", state.ID.ValueString()),
		)
		return
	}

	// Update State from Response
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *applicationCategoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.ApplicationCategoryModel
	var state model.ApplicationCategoryModel
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

	apiAppCategory := *plan.MapToApi()
	apiAppCategory.Id = state.ID.ValueStringPointer()
	apiResponse, httpResponse, err := r.Client.ApplicationCategoriesAPI.UpdateTag(
		r.AuthContext(ctx),
		plan.OrganizationId.ValueString(),
	).ApiApplicationCategoryDTO(apiAppCategory).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error updating Application Category for Organization",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.HandleAPIError(
			"Updating Application Category for Organization was not successful",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// Map response to State
	plan.MapFromApi(apiResponse)

	// Update State
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *applicationCategoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state model.ApplicationCategoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	httpResponse, err := r.Client.ApplicationCategoriesAPI.DeleteTag(
		r.AuthContext(ctx),
		state.OrganizationId.ValueString(),
		state.ID.ValueString(),
	).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			fmt.Sprintf(common.ERR_APPLICATION_CATEGORY_FOR_ORG_DID_NOT_EXIST, state.ID.ValueString()),
			fmt.Sprintf("%v", err),
		)
		return
	}
}

// Import
// Format: ORGANIZATION_ID,APPLICATION_CATEGORY_ID
func (r *applicationCategoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: <organization-id>,<applicaiton-category-id> - Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
