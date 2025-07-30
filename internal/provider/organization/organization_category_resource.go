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
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// applicationCategoryResource is the resource implementation.
type applicationCategoryResource struct {
	common.BaseResourceWithImport
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
			"id": schema.StringAttribute{
				Description: "Internal ID of the Application Category",
				Computed:    true,
				Required:    false,
				Optional:    false,
			},
			"name": schema.StringAttribute{
				Description: "Name of the Application Category",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the Application Category",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Internal ID of the Organization to which this Application Category belongs. Use `ROOT_ORGANIZATION_ID` for the Root Organization.",
				Required:    true,
			},
			"color": schema.StringAttribute{
				Description: "Color of the Application Category",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(model.AllColors()...),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *applicationCategoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan model.ApplicationCategoryModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Call API to Create
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	apiDto := sonatypeiq.NewApiApplicationCategoryDTOWithDefaults()
	plan.MapToApi(apiDto)
	apiResponse, httpResponse, err := r.Client.ApplicationCategoriesAPI.AddTag(
		ctx, plan.OrganizationId.ValueString(),
	).ApiApplicationCategoryDTO(*apiDto).Execute()

	// Handle Errors
	if err != nil {
		common.HandleApiError(
			"Error creating Application Category",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}
	if httpResponse.StatusCode != http.StatusOK {
		common.HandleApiError(
			"Creation of Application Category unsuccesful",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
	}

	plan.MapFromApi(apiResponse)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *applicationCategoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.ApplicationCategoryModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)

	// Handle any errors
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Set API Context
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Make API Request
	apiResponse, httpResponse, err := r.Client.ApplicationCategoriesAPI.GetTags(ctx, state.OrganizationId.ValueString()).Execute()

	// Handle any errors
	if err != nil {
		common.HandleApiError(
			"Unable to read Application Categories",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	if len(apiResponse) == 0 {
		resp.State.RemoveResource(ctx)
		common.HandleApiWarning(
			"No Application Categories exist",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	found := false
	for _, ac := range apiResponse {
		if *ac.Id == state.ID.ValueString() {
			state.MapFromApi(&ac)
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		common.HandleApiWarning(
			fmt.Sprintf("Application Category with ID %s does not exist", state.ID.ValueString()),
			&err,
			httpResponse,
			&resp.Diagnostics,
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
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	var state model.ApplicationCategoryModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Request Context
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Make API Request
	apiDto := sonatypeiq.NewApiApplicationCategoryDTOWithDefaults()
	plan.MapToApi(apiDto)
	apiResponse, httpResponse, err := r.Client.ApplicationCategoriesAPI.UpdateTag(
		ctx, plan.OrganizationId.ValueString(),
	).ApiApplicationCategoryDTO(*apiDto).Execute()

	// Handle Errors
	if err != nil {
		common.HandleApiError(
			"Error updating Application Category",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}
	if httpResponse.StatusCode != http.StatusOK {
		common.HandleApiError(
			"Updating Application Category unsuccesful",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
	}

	// Update State
	plan.MapFromApi(apiResponse)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *applicationCategoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.ApplicationCategoryModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Request Context
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Make API request
	httpResponse, err := r.Client.ApplicationCategoriesAPI.DeleteTag(
		ctx, state.OrganizationId.ValueString(), state.ID.ValueString(),
	).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			common.HandleApiWarning(
				"Application Category does not exist in Organization with ID in state",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		} else {
			common.HandleApiError(
				"Failed to delete Application Category",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		}
		return
	}

	if httpResponse.StatusCode != http.StatusNoContent {
		common.HandleApiError(
			"Unexpected response code deleting Application Category",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
	}
}

func (r *applicationCategoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
