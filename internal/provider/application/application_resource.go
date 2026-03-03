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
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
	sharedutil "github.com/sonatype-nexus-community/terraform-provider-shared/util"
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
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to create Application
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	application_request := r.Client.ApplicationsAPI.AddApplication(ctx)
	application_request = application_request.ApiApplicationDTO(sonatypeiq.ApiApplicationDTO{
		Name:            sharedutil.StringToPtr(plan.Name.ValueString()),
		PublicId:        sharedutil.StringToPtr(plan.PublicId.ValueString()),
		OrganizationId:  sharedutil.StringToPtr(plan.OrganizationId.ValueString()),
		ContactUserName: sharedutil.StringToPtr(plan.ContactUserName.ValueString()),
	})
	application, api_response, err := application_request.Execute()

	// Call API
	if err != nil {
		sharederr.HandleAPIError("Error creating Application", &err, api_response, &resp.Diagnostics)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = sharedutil.StringPtrToValue(application.Id)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *applicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.ApplicationModelResource

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Get refreshed Application from IQ
	application, api_response, err := r.Client.ApplicationsAPI.GetApplication(ctx, state.ID.ValueString()).Execute()

	if err != nil {
		if sharederr.IsNotFound(api_response.StatusCode) {
			resp.State.RemoveResource(ctx)
		} else {
			sharederr.HandleAPIError("Error Reading IQ Application", &err, api_response, &resp.Diagnostics)
		}
		return
	} else {
		// Overwrite items with refreshed state
		state.ID = sharedutil.StringPtrToValue(application.Id)
		state.Name = sharedutil.StringPtrToValue(application.Name)
		state.PublicId = sharedutil.StringPtrToValue(application.PublicId)
		state.OrganizationId = sharedutil.StringPtrToValue(application.OrganizationId)
		if state.LastUpdated == types.StringValue("") {
			state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
		}
		state.ContactUserName = sharedutil.StringPtrToValue(application.ContactUserName)
	}

	// Set refreshed state
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
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make Update API Call
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)
	if !plan.OrganizationId.Equal(state.OrganizationId) {
		_, apiResponse, err := r.Client.ApplicationsAPI.MoveApplication(ctx, state.ID.ValueString(), plan.OrganizationId.ValueString()).Execute()
		if err != nil {
			sharederr.HandleAPIError("Error moving application", &err, apiResponse, &resp.Diagnostics)
			return
		}
	}
	app_update_request := r.Client.ApplicationsAPI.UpdateApplication(ctx, state.ID.ValueString())
	app_update_request = app_update_request.ApiApplicationDTO(sonatypeiq.ApiApplicationDTO{
		Name:            sharedutil.StringToPtr(plan.Name.ValueString()),
		PublicId:        sharedutil.StringToPtr(plan.PublicId.ValueString()),
		OrganizationId:  sharedutil.StringToPtr(plan.OrganizationId.ValueString()),
		ContactUserName: sharedutil.StringToPtr(plan.ContactUserName.ValueString()),
	})

	application, api_response, err := app_update_request.Execute()

	// Call API
	if err != nil {
		sharederr.HandleAPIError("Error updating Application", &err, api_response, &resp.Diagnostics)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = sharedutil.StringPtrToValue(application.Id)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.ApplicationModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make Delete API Call
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	api_response, err := r.Client.ApplicationsAPI.DeleteApplication(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		sharederr.HandleAPIError("Error deleting Application", &err, api_response, &resp.Diagnostics)
		return
	}
}

func (r *applicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
