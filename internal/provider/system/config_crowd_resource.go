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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
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
			"id":                   sharedrschema.ResourceComputedString("Internal ID for Terraform State"),
			"server_url":           sharedrschema.ResourceRequiredString("Crowd Server URL"),
			"application_name":     sharedrschema.ResourceRequiredString("Crowd Application Name"),
			"application_password": sharedrschema.ResourceSensitiveRequiredString("Crowd Application Password"),
			"last_updated":         sharedrschema.ResourceLastUpdated(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *configCrowdResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.ConfigCrowdModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	r.doUpsert(ctx, &plan, &resp.Diagnostics)

	// Update State
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *configCrowdResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.ConfigCrowdModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := r.Client.ConfigCrowdAPI.GetCrowdConfiguration(r.AuthContext(ctx)).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"Crowd configuraiton does not exist",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		} else {
			errors.HandleAPIError(
				common.ERR_FAILED_READING_CROWD_CONFIGURATION,
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
func (r *configCrowdResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.ConfigCrowdModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	r.doUpsert(ctx, &plan, &resp.Diagnostics)

	// Update State
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *configCrowdResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	httpResponse, err := r.Client.ConfigCrowdAPI.DeleteCrowdConfiguration(r.AuthContext(ctx)).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			common.ERR_CROWD_CONFIGURATION_DID_NOT_EXIST,
			fmt.Sprintf("%v", err),
		)
		return
	}
}

func (r *configCrowdResource) doUpsert(ctx context.Context, model *model.ConfigCrowdModel, respDiags *diag.Diagnostics) {
	httpResponse, err := r.Client.ConfigCrowdAPI.InsertOrUpdateCrowdConfiguration(
		r.AuthContext(ctx),
	).ApiCrowdConfigurationDTO(*model.MapToApi()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating Crowd configuration",
			&err,
			httpResponse,
			respDiags,
		)
		return
	} else if httpResponse.StatusCode != http.StatusNoContent {
		errors.HandleAPIError(
			"Creation of Crowd configuration was not successful",
			&err,
			httpResponse,
			respDiags,
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	model.ID = types.StringValue(common.STATE_ID_CROWD_CONFIGURATION)
	model.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}

func (r *configCrowdResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
