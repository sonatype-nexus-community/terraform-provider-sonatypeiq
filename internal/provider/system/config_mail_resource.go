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
	sharedutil "github.com/sonatype-nexus-community/terraform-provider-shared/util"
)

const (
	mailDefaultPort            int64 = 465
	mailDefaultSSLEnabled      bool  = true
	mailDefaultStartTLSEnabled bool  = true
)

// configMailResource is the resource implementation.
type configMailResource struct {
	common.BaseResource
}

type configMailModelResource struct {
	ID                 types.String `tfsdk:"id"`
	Hostname           types.String `tfsdk:"hostname"`
	Port               types.Int64  `tfsdk:"port"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	PasswordIsIncluded types.Bool   `tfsdk:"password_is_included"`
	SSLEnabled         types.Bool   `tfsdk:"ssl_enabled"`
	StartTLSEnabled    types.Bool   `tfsdk:"start_tls_enabled"`
	SystemEmail        types.String `tfsdk:"system_email"`
	LastUpdated        types.String `tfsdk:"last_updated"`
}

// NewConfigMailResource is a helper function to simplify the provider implementation.
func NewConfigMailResource() resource.Resource {
	return &configMailResource{}
}

// Metadata returns the resource type name.
func (r *configMailResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_mail"
}

// Schema defines the schema for the resource.
func (r *configMailResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage outbound email server configuration for IQ Server",
		Attributes: map[string]schema.Attribute{
			"id":                   sharedrschema.ResourceComputedString(""),
			"hostname":             sharedrschema.ResourceRequiredString("Hostname of the SMTP server"),
			"port":                 sharedrschema.ResourceComputedOptionalInt64WithDefault("Port Number for the SMTP server", mailDefaultPort),
			"username":             sharedrschema.ResourceOptionalString("Username for the SMTP server"),
			"password":             sharedrschema.ResourceSensitiveString("Password for the SMTP server"),
			"password_is_included": sharedrschema.ResourceComputedOptionalBoolWithDefault("Whether the password is included", false),
			"ssl_enabled":          sharedrschema.ResourceComputedOptionalBoolWithDefault("Whether SSL is enabled to SMTP server", mailDefaultSSLEnabled),
			"start_tls_enabled":    sharedrschema.ResourceComputedOptionalBoolWithDefault("Whether STARTTLS is enabled to SMTP server", mailDefaultStartTLSEnabled),
			"system_email":         sharedrschema.ResourceRequiredString("The email address emails sent by Sonatype IQ Server will appear FROM"),
			"last_updated":         sharedrschema.ResourceComputedString(""),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *configMailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan configMailModelResource
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
	mail_config := sonatypeiq.ApiMailConfigurationDTO{
		Hostname:        sharedutil.StringToPtr(plan.Hostname.ValueString()),
		Port:            port,
		Username:        sharedutil.StringToPtr(plan.Username.ValueString()),
		SslEnabled:      sharedutil.BoolToPtr(plan.SSLEnabled.ValueBool()),
		StartTlsEnabled: sharedutil.BoolToPtr(plan.StartTLSEnabled.ValueBool()),
		SystemEmail:     sharedutil.StringToPtr(plan.SystemEmail.ValueString()),
	}
	if !plan.Password.IsNull() {
		mail_config.Password = sharedutil.StringToPtr(plan.Password.ValueString())
		mail_config.SetPasswordIsIncluded(true)
	} else {
		mail_config.SetPasswordIsIncluded(false)
	}

	mail_config_request := r.Client.ConfigMailAPI.SetConfiguration2(ctx)
	mail_config_request = mail_config_request.ApiMailConfigurationDTO(mail_config)
	api_response, err := mail_config_request.Execute()

	// Call API
	if err != nil {
		sharederr.HandleAPIError(
			"Error creating Mail Configuration",
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
func (r *configMailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state configMailModelResource

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Get refreshed Mail Config from IQ
	mail_config, api_response, err := r.Client.ConfigMailAPI.GetConfiguration2(ctx).Execute()

	if err != nil {
		if sharederr.IsNotFound(api_response.StatusCode) {
			resp.State.RemoveResource(ctx)
		} else {
			sharederr.HandleAPIError(
				"Error Reading IQ Mail Configuration",
				&err,
				api_response,
				&resp.Diagnostics,
			)
		}
		return
	}

	// Overwrite items with refreshed state
	state.Hostname = sharedutil.StringPtrToValue(mail_config.Hostname)
	state.Port = types.Int64Value(int64(*mail_config.Port))
	state.Username = types.StringNull()
	if mail_config.HasUsername() {
		state.Username = sharedutil.StringPtrToValue(mail_config.Username)
	}
	state.Password = types.StringNull()
	state.PasswordIsIncluded = sharedutil.BoolPtrToValue(mail_config.PasswordIsIncluded)
	state.SSLEnabled = sharedutil.BoolPtrToValue(mail_config.SslEnabled)
	state.StartTLSEnabled = sharedutil.BoolPtrToValue(mail_config.StartTlsEnabled)
	state.SystemEmail = sharedutil.StringPtrToValue(mail_config.SystemEmail)

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *configMailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan configMailModelResource
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
	mail_config := sonatypeiq.ApiMailConfigurationDTO{
		Hostname:        sharedutil.StringToPtr(plan.Hostname.ValueString()),
		Port:            port,
		Username:        sharedutil.StringToPtr(plan.Username.ValueString()),
		SslEnabled:      sharedutil.BoolToPtr(plan.SSLEnabled.ValueBool()),
		StartTlsEnabled: sharedutil.BoolToPtr(plan.StartTLSEnabled.ValueBool()),
		SystemEmail:     sharedutil.StringToPtr(plan.SystemEmail.ValueString()),
	}
	if !plan.Password.IsNull() {
		mail_config.Password = sharedutil.StringToPtr(plan.Password.ValueString())
		mail_config.SetPasswordIsIncluded(true)
	} else {
		mail_config.SetPasswordIsIncluded(false)
	}

	mail_config_request := r.Client.ConfigMailAPI.SetConfiguration2(ctx)
	mail_config_request = mail_config_request.ApiMailConfigurationDTO(mail_config)
	api_response, err := mail_config_request.Execute()

	// Call API
	if err != nil {
		sharederr.HandleAPIError(
			"Error updating Mail Configuration",
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

// Delete deletes the resource and removes the Terraform state on success.
func (r *configMailResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Make Delete API Call
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	api_response, err := r.Client.ConfigMailAPI.DeleteConfiguration2(ctx).Execute()

	if err != nil {
		sharederr.HandleAPIError(
			"Error deleting Mail Configuration",
			&err,
			api_response,
			&resp.Diagnostics,
		)
		return
	}
}
