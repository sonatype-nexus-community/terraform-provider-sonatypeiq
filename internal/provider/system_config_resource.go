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
	"io"
	"math/rand"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// systemConfigResource is the resource implementation.
type systemConfigResource struct {
	baseResource
}

type systemConfigModelResource struct {
	ID           types.String `tfsdk:"id"`
	BaseURL      types.String `tfsdk:"base_url"`
	ForceBaseURL types.Bool   `tfsdk:"force_base_url"`
	LastUpdated  types.String `tfsdk:"last_updated"`
}

// NewSystemConfigResource is a helper function to simplify the provider implementation.
func NewSystemConfigResource() resource.Resource {
	return &systemConfigResource{}
}

// Metadata returns the resource type name.
func (r *systemConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_config"
}

// Schema defines the schema for the resource.
func (r *systemConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to manage System Configuration",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"base_url": schema.StringAttribute{
				Description: "Base URL for Sonatype IQ Server. See https://help.sonatype.com/en/configuration-rest-api.html#base-url--required-",
				Required:    true,
			},
			"force_base_url": schema.BoolAttribute{
				Description: "Should the Base URL be forced? See https://help.sonatype.com/en/configuration-rest-api.html#force-the-base-url",
				Required:    true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *systemConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan systemConfigModelResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to create Organization
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.auth,
	)

	config_request := r.client.ConfigurationAPI.SetConfiguration(ctx)
	system_config := sonatypeiq.SystemConfig{}
	if !plan.BaseURL.IsNull() {
		system_config.BaseUrl = *sonatypeiq.NewNullableString(plan.BaseURL.ValueStringPointer())
	}
	if !plan.ForceBaseURL.IsNull() {
		system_config.ForceBaseUrl = *sonatypeiq.NewNullableBool(plan.ForceBaseURL.ValueBoolPointer())
	}
	config_request = config_request.SystemConfig(system_config)
	api_response, err := config_request.Execute()

	// Call API
	if err != nil || api_response.StatusCode != 204 {
		error_body, _ := io.ReadAll(api_response.Body)
		resp.Diagnostics.AddError(
			"Error creating System Configuration",
			"Could not create System Configuration, unexpected error: "+api_response.Status+": "+string(error_body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(strconv.FormatUint(uint64(rand.Uint32())<<32+uint64(rand.Uint32()), 36))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *systemConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state systemConfigModelResource

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.auth,
	)

	// Lookup System Configuration
	config_request := r.client.ConfigurationAPI.GetConfiguration(ctx)
	config_request = config_request.Property([]sonatypeiq.SystemConfigProperty{
		"baseUrl", "forceBaseUrl",
	})
	system_config, api_response, err := config_request.Execute()

	if err != nil || api_response.StatusCode != 200 {
		error_body, _ := io.ReadAll(api_response.Body)
		resp.Diagnostics.AddError(
			"Error reading System Configuration",
			"Could not read System Configuration, unexpected error: "+api_response.Status+": "+string(error_body),
		)
		return
	}

	print(system_config)

	if system_config.BaseUrl.IsSet() {
		state.BaseURL = types.StringValue(system_config.GetBaseUrl())
	}
	if system_config.ForceBaseUrl.IsSet() {
		state.ForceBaseURL = types.BoolValue(system_config.GetForceBaseUrl())
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *systemConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan systemConfigModelResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to create Organization
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.auth,
	)

	config_request := r.client.ConfigurationAPI.SetConfiguration(ctx)
	system_config := sonatypeiq.SystemConfig{}
	if !plan.BaseURL.IsNull() {
		system_config.BaseUrl = *sonatypeiq.NewNullableString(plan.BaseURL.ValueStringPointer())
	}
	if !plan.ForceBaseURL.IsNull() {
		system_config.ForceBaseUrl = *sonatypeiq.NewNullableBool(plan.ForceBaseURL.ValueBoolPointer())
	}
	config_request = config_request.SystemConfig(system_config)
	api_response, err := config_request.Execute()

	// Call API
	if err != nil || api_response.StatusCode != 204 {
		error_body, _ := io.ReadAll(api_response.Body)
		resp.Diagnostics.AddError(
			"Error updating System Configuration",
			"Could not update System Configuration, unexpected error: "+api_response.Status+": "+string(error_body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(strconv.FormatUint(uint64(rand.Uint32())<<32+uint64(rand.Uint32()), 36))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *systemConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
