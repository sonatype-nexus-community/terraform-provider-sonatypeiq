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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// configProxyServerResource is the resource implementation.
type configProxyServerResource struct {
	common.BaseResource
}

// NewConfigProxyServerResource is a helper function to simplify the provider implementation.
func NewConfigProxyServerResource() resource.Resource {
	return &configProxyServerResource{}
}

// Metadata returns the resource type name.
func (r *configProxyServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_proxy_server"
}

// Schema defines the schema for the resource.
func (r *configProxyServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage outbound Proxy Server configuration for IQ Server",
		Attributes: map[string]schema.Attribute{
			"id":            sharedrschema.ResourceComputedString("Internal ID for Terraform State"),
			"hostname":      sharedrschema.ResourceRequiredString("Hostname of the Proxy Server"),
			"port":          sharedrschema.ResourceRequiredInt32("Port Number for the Proxy Server"),
			"username":      sharedrschema.ResourceOptionalString("Username for the Proxy Server"),
			"password":      sharedrschema.ResourceSensitiveString("Password for the Proxy Server"),
			"exclude_hosts": sharedrschema.ResourceComputedOptionalStringSet("Optional list of hosts to exclude communication via Proxy Server"),
			"last_updated":  sharedrschema.ResourceLastUpdated(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *configProxyServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.ConfigProxyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	r.doUpsert(ctx, &plan, &resp.State, &resp.Diagnostics)

	// Update State
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *configProxyServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.ConfigProxyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	apiResponse := r.doRead(ctx, &resp.State, &resp.Diagnostics)
	if apiResponse == nil {
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
func (r *configProxyServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.ConfigProxyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	r.doUpsert(ctx, &plan, &resp.State, &resp.Diagnostics)

	// Update State
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *configProxyServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	httpResponse, err := r.Client.ConfigProxyServerAPI.DeleteConfiguration3(r.AuthContext(ctx)).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			common.ERR_PROXY_CONFIGURATION_DID_NOT_EXIST,
			fmt.Sprintf("%v", err),
		)
		return
	}
}

func (r *configProxyServerResource) doRead(ctx context.Context, respState *tfsdk.State, respDiags *diag.Diagnostics) *sonatypeiq.ApiProxyServerConfigurationDTO {
	apiResponse, httpResponse, err := r.Client.ConfigProxyServerAPI.GetConfiguration3(r.AuthContext(ctx)).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			respState.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"Proxy Server configuration did not exist",
				&err,
				httpResponse,
				respDiags,
			)
		} else {
			errors.HandleAPIError(
				common.ERR_FAILED_READING_PROXY_CONFIGURATION,
				&err,
				httpResponse,
				respDiags,
			)
		}
		return nil
	}

	return apiResponse
}

func (r *configProxyServerResource) doUpsert(ctx context.Context, model *model.ConfigProxyModel, respState *tfsdk.State, respDiags *diag.Diagnostics) {
	httpResponse, err := r.Client.ConfigProxyServerAPI.SetConfiguration3(r.AuthContext(ctx)).ApiProxyServerConfigurationDTO(*model.MapToApi()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating/updating Proxy Server configuration",
			&err,
			httpResponse,
			respDiags,
		)
		return
	} else if httpResponse.StatusCode != http.StatusNoContent {
		errors.HandleAPIError(
			"Upsertion of Proxy Server configuration was not successful",
			&err,
			httpResponse,
			respDiags,
		)
		return
	}

	apiResponse := r.doRead(ctx, respState, respDiags)
	if apiResponse == nil {
		return
	}

	// Map response to State
	model.MapFromApi(apiResponse)
	model.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}

func (r *configProxyServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
