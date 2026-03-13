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

// configMailResource is the resource implementation.
type configMailResource struct {
	common.BaseResource
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
			"id":                sharedrschema.ResourceComputedString("Internal ID for Terraform State"),
			"hostname":          sharedrschema.ResourceRequiredString("Hostname of the SMTP server"),
			"port":              sharedrschema.ResourceComputedOptionalInt32WithDefault("Port Number for the SMTP server", common.DEFAULT_MAIL_SERVER_PORT),
			"username":          sharedrschema.ResourceOptionalString("Username for the SMTP server"),
			"password":          sharedrschema.ResourceSensitiveString("Password for the SMTP server"),
			"ssl_enabled":       sharedrschema.ResourceComputedOptionalBoolWithDefault("Whether SSL is enabled to SMTP server", common.DEFAULT_MAIL_SSL_ENABLED),
			"start_tls_enabled": sharedrschema.ResourceComputedOptionalBoolWithDefault("Whether STARTTLS is enabled to SMTP server", common.DEFAULT_MAIL_START_TLS_ENABLED),
			"system_email":      sharedrschema.ResourceRequiredString("The email address emails sent by Sonatype IQ Server will appear FROM"),
			"last_updated":      sharedrschema.ResourceLastUpdated(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *configMailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.ConfigMailModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

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
func (r *configMailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.ConfigMailModel
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
func (r *configMailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.ConfigMailModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

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
func (r *configMailResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	httpResponse, err := r.Client.ConfigMailAPI.DeleteConfiguration2(r.AuthContext(ctx)).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			common.ERR_MAIL_CONFIGURATION_DID_NOT_EXIST,
			fmt.Sprintf("%v", err),
		)
		return
	}
}

func (r *configMailResource) doRead(ctx context.Context, respState *tfsdk.State, respDiags *diag.Diagnostics) *sonatypeiq.ApiMailConfigurationDTO {
	apiResponse, httpResponse, err := r.Client.ConfigMailAPI.GetConfiguration2(r.AuthContext(ctx)).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			respState.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"Mail configuration did not exist",
				&err,
				httpResponse,
				respDiags,
			)
		} else {
			errors.HandleAPIError(
				common.ERR_FAILED_READING_MAIL_CONFIGURATION,
				&err,
				httpResponse,
				respDiags,
			)
		}
		return nil
	}

	return apiResponse
}

func (r *configMailResource) doUpsert(ctx context.Context, model *model.ConfigMailModel, respState *tfsdk.State, respDiags *diag.Diagnostics) {
	httpResponse, err := r.Client.ConfigMailAPI.SetConfiguration2(r.AuthContext(ctx)).ApiMailConfigurationDTO(*model.MapToApi()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating/updating Mail configuration",
			&err,
			httpResponse,
			respDiags,
		)
		return
	} else if httpResponse.StatusCode != http.StatusNoContent {
		errors.HandleAPIError(
			"Upsertion of Mail configuration was not successful",
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
	model.ID = types.StringValue(common.STATE_ID_MAIL_CONFIGURATION)
	model.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}

func (r *configMailResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
