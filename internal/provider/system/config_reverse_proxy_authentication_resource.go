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

// configReverseProxyAuthenticationResource is the resource implementation.
type configReverseProxyAuthenticationResource struct {
	common.BaseResource
}

// NewConfigReverseProxyAuthenticationResource is a helper function to simplify the provider implementation.
func NewConfigReverseProxyAuthenticationResource() resource.Resource {
	return &configReverseProxyAuthenticationResource{}
}

// Metadata returns the resource type name.
func (r *configReverseProxyAuthenticationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_reverse_proxy_authentication"
}

// Schema defines the schema for the resource.
func (r *configReverseProxyAuthenticationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Reverse Proxy Authentication configuration for IQ Server",
		Attributes: map[string]schema.Attribute{
			"id": sharedrschema.ResourceComputedString("Internal ID for Terraform State"),
			"enabled": schema.BoolAttribute{
				Description: "Whether reverse proxy authentication is enabled. When enabled, users authenticate via a " +
					"reverse proxy that provides the username in an HTTP header. This changes how users sign in and may " +
					"affect the account Terraform itself connects as - ensure you understand the implications before enabling.",
				Required: true,
			},
			"username_header": schema.StringAttribute{
				Description: "The name of the HTTP request header that contains the username. " +
					"When not set, defaults to 'REMOTE_USER' on the server.",
				Optional: true,
			},
			"csrf_protection_disabled": schema.BoolAttribute{
				Description: "Whether Cross-Site Request Forgery (CSRF) protection is disabled. " +
					"Used for backward compatibility with old client plugins. Disabling CSRF protection is a " +
					"security risk - only set this to true if absolutely necessary for compatibility with legacy clients.",
				Optional: true,
			},
			"logout_url": schema.StringAttribute{
				Description: "The redirect URL when a user logs out. Leave unset to prevent redirection on logout.",
				Optional:    true,
			},
			"last_updated": sharedrschema.ResourceLastUpdated(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *configReverseProxyAuthenticationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.ConfigReverseProxyAuthenticationModel
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
func (r *configReverseProxyAuthenticationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.ConfigReverseProxyAuthenticationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	apiResponse := r.doRead(ctx, &resp.State, &resp.Diagnostics)
	if apiResponse == nil {
		return
	}

	// Update State based on Response, preserving null for unset optional attributes
	state.MapFromApi(apiResponse, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *configReverseProxyAuthenticationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state model.ConfigReverseProxyAuthenticationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Special handling for username_header removal:
	// The server does not clear usernameHeader when it's absent from PUT.
	// When going from set (state) to null (plan), DELETE first to reset to defaults.
	if !state.UsernameHeader.IsNull() && plan.UsernameHeader.IsNull() {
		httpResponse, err := r.Client.ConfigReverseProxyAuthenticationAPI.DeleteConfiguration5(r.AuthContext(ctx)).Execute()
		if err != nil || httpResponse.StatusCode != http.StatusNoContent {
			resp.Diagnostics.AddError(
				"Error resetting configuration before update",
				fmt.Sprintf("Failed to delete configuration to reset username_header: %v", err),
			)
			return
		}
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
func (r *configReverseProxyAuthenticationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	httpResponse, err := r.Client.ConfigReverseProxyAuthenticationAPI.DeleteConfiguration5(r.AuthContext(ctx)).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			common.ERR_REVERSE_PROXY_AUTHENTICATION_CONFIGURATION_DID_NOT_EXIST,
			fmt.Sprintf("%v", err),
		)
		return
	}
}

func (r *configReverseProxyAuthenticationResource) doRead(ctx context.Context, respState *tfsdk.State, respDiags *diag.Diagnostics) *sonatypeiq.ApiReverseProxyAuthenticationConfigurationDTO {
	apiResponse, httpResponse, err := r.Client.ConfigReverseProxyAuthenticationAPI.GetConfiguration5(r.AuthContext(ctx)).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			respState.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"Reverse proxy authentication configuration did not exist",
				&err,
				httpResponse,
				respDiags,
			)
		} else {
			errors.HandleAPIError(
				common.ERR_FAILED_READING_REVERSE_PROXY_AUTHENTICATION_CONFIGURATION,
				&err,
				httpResponse,
				respDiags,
			)
		}
		return nil
	}

	return apiResponse
}

func (r *configReverseProxyAuthenticationResource) doUpsert(ctx context.Context, model *model.ConfigReverseProxyAuthenticationModel, respState *tfsdk.State, respDiags *diag.Diagnostics) {
	httpResponse, err := r.Client.ConfigReverseProxyAuthenticationAPI.SetConfiguration5(r.AuthContext(ctx)).ApiReverseProxyAuthenticationConfigurationDTO(*model.MapToApi()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating/updating reverse proxy authentication configuration",
			&err,
			httpResponse,
			respDiags,
		)
		return
	} else if httpResponse.StatusCode != http.StatusNoContent {
		errors.HandleAPIError(
			"Upsertion of reverse proxy authentication configuration was not successful",
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

	// Map response to State, preserving null for attributes not explicitly set in the plan
	model.MapFromApi(apiResponse, model)
	model.ID = types.StringValue(common.STATE_ID_REVERSE_PROXY_AUTHENTICATION_CONFIGURATION)
	model.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}

func (r *configReverseProxyAuthenticationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
