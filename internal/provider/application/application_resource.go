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
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

var _ resource.ResourceWithImportState = &applicationResource{}

// applicationResource is the resource implementation.
type applicationResource struct {
	common.BaseResource
}

// NewApplicationResource is a helper function to simplify the provider implementation.
func NewApplicationResource() resource.Resource {
	return &applicationResource{}
}

// Metadata returns the resource type name.
func (r *applicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

// Schema defines the schema for the resource.
func (r *applicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                sharedrschema.ResourceComputedString("Internal ID of the Application"),
			"name":              sharedrschema.ResourceRequiredString("Name of the Application"),
			"public_id":         sharedrschema.ResourceRequiredString("Public ID of the Application"),
			"organization_id":   sharedrschema.ResourceRequiredString("Internal ID of the Organization to which this Application belongs"),
			"contact_user_name": sharedrschema.ResourceOptionalString("User Name of the Contact for the Application"),
			"last_updated":      sharedrschema.ResourceLastUpdated(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *applicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan model.ApplicationModelResource
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := r.Client.ApplicationsAPI.AddApplication(r.AuthContext(ctx)).ApiApplicationDTO(plan.MapToApi()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating Application",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.HandleAPIError(
			"Creation of Application was not successful",
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
func (r *applicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.ApplicationModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := r.Client.ApplicationsAPI.GetApplication(r.AuthContext(ctx), state.ID.ValueString()).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"Application with ID did not exist",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		} else {
			errors.HandleAPIError(
				common.ERR_FAILED_READING_APPLICATION,
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		}
		return
	}

	// Update State based on Response
	state.MapFromApi(apiResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *applicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.ApplicationModelResource
	var state model.ApplicationModelResource
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

	// First Move Application if Org is different
	if !plan.OrganizationId.Equal(state.OrganizationId) {
		_, httpResponse, err := r.Client.ApplicationsAPI.MoveApplication(
			r.AuthContext(ctx), state.ID.ValueString(), plan.OrganizationId.ValueString(),
		).Execute()

		if err != nil {
			errors.HandleAPIError(
				common.ERR_FAILED_MOVING_APPLICATION,
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
			return
		}
	}

	// Second Update Application
	apiResponse, httpResponse, err := r.Client.ApplicationsAPI.UpdateApplication(r.AuthContext(ctx), state.ID.ValueString()).ApiApplicationDTO(plan.MapToApi()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error updating Application",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.HandleAPIError(
			"Updating Application was not successful",
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
func (r *applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state model.ApplicationModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	httpResponse, err := r.Client.ApplicationsAPI.DeleteApplication(r.AuthContext(ctx), state.ID.ValueString()).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			fmt.Sprintf(common.ERR_APPLICATION_DID_NOT_EXIST, state.ID.ValueString()),
			fmt.Sprintf("%v", err),
		)
		return
	}
}

func (r *applicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
