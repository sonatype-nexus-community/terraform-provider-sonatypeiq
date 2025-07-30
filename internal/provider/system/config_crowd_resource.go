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

package system

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// configCrowdResource is the resource implementation.
type configCrowdResource struct {
	common.BaseResource
}

// NewConfigCrowdResource is a helper function to simplify the provider implementation.
func NewConfigCrowdResource() resource.Resource {
	return &configCrowdResource{}
}

// Metadata returns the resource type name.
func (r *configCrowdResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_crowd"
}

// Schema defines the schema for the resource.
func (r *configCrowdResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Atlassian Crowd server configuration for IQ Server",
		Attributes: map[string]schema.Attribute{
			"server_url": schema.StringAttribute{
				Description: "Crowd Server URL",
				Required:    true,
			},
			"application_name": schema.StringAttribute{
				Description: "Crowd Application Name",
				Required:    true,
			},
			"application_password": schema.StringAttribute{
				Description: "Crowd Application Password",
				Required:    true,
				Sensitive:   true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *configCrowdResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan model.ConfigCrowdModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.doUpsert(ctx, &plan, &resp.Diagnostics)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *configCrowdResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.ConfigCrowdModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

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

	apiResponse, httpResponse, err := r.Client.ConfigCrowdAPI.GetCrowdConfiguration(ctx).Execute()

	// Handle any errors
	if err != nil && httpResponse.StatusCode != http.StatusNotFound {
		common.HandleApiError(
			"Unable to read Crowd configuration",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// Set refreshed state
	state.MapFromApi(apiResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *configCrowdResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan model.ConfigCrowdModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.doUpsert(ctx, &plan, &resp.Diagnostics)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *configCrowdResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Make Delete API Call
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	httpResponse, err := r.Client.ConfigCrowdAPI.DeleteCrowdConfiguration(ctx).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			common.HandleApiWarning(
				"There is no current Crowd configuration",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		} else {
			common.HandleApiError(
				"Failed to delete Crowd configuration",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		}
		return
	}

	if httpResponse.StatusCode != http.StatusNoContent {
		common.HandleApiError(
			"Unexpected response code deleting Crowd configuration",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
	}
}

func (r *configCrowdResource) doUpsert(ctx context.Context, model *model.ConfigCrowdModel, respDiags *diag.Diagnostics) {
	// Set API context
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Call API
	apiDto := sonatypeiq.NewApiCrowdConfigurationDTOWithDefaults()
	model.MapToApi(apiDto)
	httpResponse, err := r.Client.ConfigCrowdAPI.InsertOrUpdateCrowdConfiguration(ctx).ApiCrowdConfigurationDTO(*apiDto).Execute()

	// Handle Errors
	if err != nil {
		common.HandleApiError(
			"Error creating Crowd configuration",
			&err,
			httpResponse,
			respDiags,
		)
		return
	}
	if httpResponse.StatusCode != http.StatusNoContent {
		common.HandleApiError(
			"Creation of Crowd configuration unsuccesful",
			&err,
			httpResponse,
			respDiags,
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	model.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}
