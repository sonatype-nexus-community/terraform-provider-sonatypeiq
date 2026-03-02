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
	"math/rand"
	"strconv"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// configProxyServerResource is the resource implementation.
type configProxyServerResource struct {
	common.BaseResource
}

type configProxyServerModelResource struct {
	ID                 types.String `tfsdk:"id"`
	Hostname           types.String `tfsdk:"hostname"`
	Port               types.Int64  `tfsdk:"port"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	PasswordIsIncluded types.Bool   `tfsdk:"password_is_included"`
	ExcludeHosts       types.Set    `tfsdk:"exclude_hosts"`
	LastUpdated        types.String `tfsdk:"last_updated"`
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
		Description: "Manage outbound proxy server configuration for IQ Server",
		Attributes: map[string]schema.Attribute{
			"id":                   sharedrschema.ResourceComputedString(""),
			"hostname":             sharedrschema.ResourceRequiredString("Hostname of the Proxy Server"),
			"port":                 sharedrschema.ResourceRequiredInt64("Port Number for the Proxy Server"),
			"username":             sharedrschema.ResourceOptionalString("Username for the Proxy Server"),
			"password":             sharedrschema.ResourceSensitiveString("Password for the Proxy Server"),
			"password_is_included": sharedrschema.ResourceComputedOptionalBoolWithDefault("Whether the password is included", false),
			"exclude_hosts":        sharedrschema.ResourceComputedOptionalStringSet("Optional list of hosts to exclude communication via Proxy Server"),
			"last_updated":         sharedrschema.ResourceComputedString(""),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *configProxyServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan configProxyServerModelResource
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

	var port = new(int32)
	*port = int32(plan.Port.ValueInt64())
	proxy_config := sonatypeiq.ApiProxyServerConfigurationDTO{
		Hostname: plan.Hostname.ValueStringPointer(),
		Port:     port,
		Username: plan.Username.ValueStringPointer(),
	}
	if !plan.Password.IsNull() {
		proxy_config.Password = plan.Password.ValueStringPointer()
		proxy_config.SetPasswordIsIncluded(true)
	} else {
		proxy_config.SetPasswordIsIncluded(false)
	}

	for _, exclude_host := range plan.ExcludeHosts.Elements() {
		proxy_config.ExcludeHosts = append(proxy_config.ExcludeHosts, exclude_host.String())
	}

	proxy_config_request := r.Client.ConfigProxyServerAPI.SetConfiguration3(ctx)
	proxy_config_request = proxy_config_request.ApiProxyServerConfigurationDTO(proxy_config)
	api_response, err := proxy_config_request.Execute()

	// Call API
	if err != nil {
		sharederr.HandleAPIError(
			"Error creating Proxy Server Configuration",
			&err,
			api_response,
			&resp.Diagnostics,
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
func (r *configProxyServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state configProxyServerModelResource

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Get refreshed Proxy Server Config from IQ
	proxy_config, api_response, err := r.Client.ConfigProxyServerAPI.GetConfiguration3(ctx).Execute()

	if err != nil {
		if sharederr.IsNotFound(api_response.StatusCode) {
			resp.State.RemoveResource(ctx)
		} else {
			sharederr.HandleAPIError(
				"Error Reading IQ Proxy Server Configuration",
				&err,
				api_response,
				&resp.Diagnostics,
			)
		}
		return
	}

	// Overwrite items with refreshed state
	state.Hostname = types.StringValue(*proxy_config.Hostname)
	state.Port = types.Int64Value(int64(*proxy_config.Port))
	state.Username = types.StringValue(*proxy_config.Username)
	state.Password = types.StringNull()
	state.PasswordIsIncluded = types.BoolValue(*proxy_config.PasswordIsIncluded)
	state.ExcludeHosts, _ = types.SetValueFrom(ctx, types.StringType, proxy_config.ExcludeHosts)

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *configProxyServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan configProxyServerModelResource
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

	var port = new(int32)
	*port = int32(plan.Port.ValueInt64())
	proxy_config := sonatypeiq.ApiProxyServerConfigurationDTO{
		Hostname: plan.Hostname.ValueStringPointer(),
		Port:     port,
		Username: plan.Username.ValueStringPointer(),
	}
	if !plan.Password.IsNull() {
		proxy_config.Password = plan.Password.ValueStringPointer()
		proxy_config.SetPasswordIsIncluded(true)
	} else {
		proxy_config.SetPasswordIsIncluded(false)
	}

	for _, exclude_host := range plan.ExcludeHosts.Elements() {
		proxy_config.ExcludeHosts = append(proxy_config.ExcludeHosts, exclude_host.String())
	}

	proxy_config_request := r.Client.ConfigProxyServerAPI.SetConfiguration3(ctx)
	proxy_config_request = proxy_config_request.ApiProxyServerConfigurationDTO(proxy_config)
	api_response, err := proxy_config_request.Execute()

	// Call API
	if err != nil {
		sharederr.HandleAPIError(
			"Error updating Proxy Server Configuration",
			&err,
			api_response,
			&resp.Diagnostics,
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *configProxyServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Make Delete API Call
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	api_response, err := r.Client.ConfigProxyServerAPI.DeleteConfiguration3(ctx).Execute()

	if err != nil {
		sharederr.HandleAPIError(
			"Error deleting Proxy Server Configuration",
			&err,
			api_response,
			&resp.Diagnostics,
		)
		return
	}
}
