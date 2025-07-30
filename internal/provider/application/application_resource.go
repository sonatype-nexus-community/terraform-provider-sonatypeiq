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
	"io"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
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
			"id": schema.StringAttribute{
				Description: "Internal ID of the Application",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the Application",
				Required:    true,
			},
			"public_id": schema.StringAttribute{
				Description: "Public ID of the Application",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Internal ID of the Organization to which this Application belongs",
				Required:    true,
			},
			"contact_user_name": schema.StringAttribute{
				Description: "User Name of the Contact for the Application",
				Optional:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "String representation of the date/time the resource was last changed by Terraform",
				Computed:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *applicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan model.ApplicationModellResource
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
		Name:            plan.Name.ValueStringPointer(),
		PublicId:        plan.PublicId.ValueStringPointer(),
		OrganizationId:  plan.OrganizationId.ValueStringPointer(),
		ContactUserName: plan.ContactUserName.ValueStringPointer(),
	})
	application, api_response, err := application_request.Execute()

	// Call API
	if err != nil {
		error_body, _ := io.ReadAll(api_response.Body)
		resp.Diagnostics.AddError(
			"Error creating Application",
			"Could not create Application, unexpected error: "+api_response.Status+": "+string(error_body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(*application.Id)
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
	var state model.ApplicationModellResource

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
		if api_response.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error Reading IQ Application",
				"Could not read Application with ID "+state.ID.ValueString()+": "+err.Error(),
			)
		}
		return
	} else {
		// Overwrite items with refreshed state
		state.ID = types.StringValue(*application.Id)
		state.Name = types.StringValue(*application.Name)
		state.PublicId = types.StringValue(*application.PublicId)
		state.OrganizationId = types.StringValue(*application.OrganizationId)
		if state.LastUpdated == types.StringValue("") {
			state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
		}
		if application.ContactUserName != nil {
			state.ContactUserName = types.StringValue(*application.ContactUserName)
		} else {
			state.ContactUserName = types.StringNull()
		}
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *applicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.ApplicationModellResource
	var state model.ApplicationModellResource
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
			errorBody, _ := io.ReadAll(apiResponse.Body)
			resp.Diagnostics.AddError(
				"Error moving application", "Could not move the application("+state.ID.ValueString()+") to new organization("+plan.OrganizationId.String()+"): "+apiResponse.Status+": "+string(errorBody),
			)
			return
		}
	}
	app_update_request := r.Client.ApplicationsAPI.UpdateApplication(ctx, state.ID.ValueString())
	app_update_request = app_update_request.ApiApplicationDTO(sonatypeiq.ApiApplicationDTO{
		Name:            plan.Name.ValueStringPointer(),
		PublicId:        plan.PublicId.ValueStringPointer(),
		OrganizationId:  plan.OrganizationId.ValueStringPointer(),
		ContactUserName: plan.ContactUserName.ValueStringPointer(),
	})

	application, api_response, err := app_update_request.Execute()

	// Call API
	if err != nil {
		error_body, _ := io.ReadAll(api_response.Body)
		resp.Diagnostics.AddError(
			"Error updating Application",
			"Could not update Application, unexpected error: "+api_response.Status+": "+string(error_body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(*application.Id)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.ApplicationModellResource
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
		error_body, _ := io.ReadAll(api_response.Body)
		resp.Diagnostics.AddError(
			"Error deleting Application",
			"Could not delete Application, unexpected error: "+api_response.Status+": "+string(error_body),
		)
		return
	}
}

func (r *applicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
