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

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// configOidcResource is the resource implementation.
type configOidcResource struct {
	common.BaseResource
}

// oidcIntroducedIn is the Sonatype IQ Server release that introduced the OIDC configuration endpoint.
// The server's own release history says 198, but endpoints have turned out to be available one release later
// than that, so confirm it against the licensed test matrix.
const oidcIntroducedIn int32 = 198

// ModifyPlan fails the plan when the connected Sonatype IQ Server is too old to support OIDC.
func (r *configOidcResource) ModifyPlan(_ context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Nothing to check when destroying
	if req.Plan.Raw.IsNull() {
		return
	}

	common.RequireIqRelease(&resp.Diagnostics, r.IqVersion, oidcIntroducedIn, "The sonatypeiq_config_oidc resource")
}

// ConfigValidators makes the plan fail for what the server would reject: it needs a JWKS URL or a JWKS.
func (r *configOidcResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("oauth2_idp_jwks_url"),
			path.MatchRoot("oauth2_idp_jwks"),
		),
	}
}

// NewConfigOidcResource is a helper function to simplify the provider implementation.
func NewConfigOidcResource() resource.Resource {
	return &configOidcResource{}
}

// Metadata returns the resource type name.
func (r *configOidcResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_oidc"
}

// Schema defines the schema for the resource.
func (r *configOidcResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Configure Sonatype IQ OIDC Single Sign-On. "+
			"~> **Note** OIDC SSO configurations will affect how all users sign in to Sonatype IQ Server. "+
			"This resource is only supported from Sonatype IQ Server release %d onwards.", oidcIntroducedIn),
		Attributes: map[string]schema.Attribute{
			"id": sharedrschema.ResourceComputedString("Internal ID for Terraform State"),
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The OAuth 2.0 client identifier for the IQ Server.",
				Required:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The OAuth 2.0 client secret for the IQ Server. This attribute is sensitive and is not returned by the server on read.",
				Required:            true,
				Sensitive:           true,
			},
			"idp_issuer": schema.StringAttribute{
				MarkdownDescription: "The issuer identifier for the Identity Provider.",
				Required:            true,
			},
			"idp_authorization_url": schema.StringAttribute{
				MarkdownDescription: "The authorization endpoint URL for the Identity Provider.",
				Required:            true,
			},
			"idp_token_url": schema.StringAttribute{
				MarkdownDescription: "The token endpoint URL for the Identity Provider.",
				Required:            true,
			},
			"authorization_custom_params_json": schema.StringAttribute{
				MarkdownDescription: "Custom parameters to include in the authorization request, as a JSON object string.",
				Optional:            true,
			},
			"token_request_custom_params_json": schema.StringAttribute{
				MarkdownDescription: "Custom parameters to include in the token request, as a JSON object string.",
				Optional:            true,
			},
			"oauth2_idp_jwks_url": schema.StringAttribute{
				MarkdownDescription: "The JWKS endpoint URL for the Identity Provider. At least one of this or `oauth2_idp_jwks` must be set.",
				Optional:            true,
			},
			"oauth2_idp_jwks": schema.StringAttribute{
				MarkdownDescription: "The JWKS (JSON Web Key Set) for the Identity Provider, as a JSON string. At least one of this or `oauth2_idp_jwks_url` must be set.",
				Optional:            true,
				Sensitive:           true,
			},
			"oauth2_idp_jws_algorithm": schema.StringAttribute{
				MarkdownDescription: "The JWS (JSON Web Signature) algorithm used to verify tokens, for example `RS256`. `none` is not allowed.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.NoneOfCaseInsensitive("none"),
				},
			},
			"oauth2_expected_audience": schema.StringAttribute{
				MarkdownDescription: "The expected audience for the JWT tokens.",
				Optional:            true,
			},
			"oauth2_username_claim": schema.StringAttribute{
				MarkdownDescription: "The JWT claim to use for the username.",
				Optional:            true,
			},
			"oauth2_email_claim": schema.StringAttribute{
				MarkdownDescription: "The JWT claim to use for the email address.",
				Optional:            true,
			},
			"oauth2_first_name_claim": schema.StringAttribute{
				MarkdownDescription: "The JWT claim to use for the first name.",
				Optional:            true,
			},
			"oauth2_last_name_claim": schema.StringAttribute{
				MarkdownDescription: "The JWT claim to use for the last name.",
				Optional:            true,
			},
			"oauth2_groups_claim": schema.StringAttribute{
				MarkdownDescription: "The JWT claim to use for groups.",
				Optional:            true,
			},
			"oauth2_exact_match_claims_json": schema.StringAttribute{
				MarkdownDescription: "Exact match claims for authorization, as a JSON object string.",
				Optional:            true,
			},
			"last_updated": sharedrschema.ResourceLastUpdated(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *configOidcResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.ConfigOidcModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
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
func (r *configOidcResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.ConfigOidcModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	// Preserve the configured secret - server returns "*******" mask
	configuredSecret := state.ClientSecret.ValueString()

	apiResponse := r.doRead(ctx, &resp.State, &resp.Diagnostics)
	if apiResponse == nil {
		return
	}

	// Update State based on Response
	state.MapFromApi(apiResponse, configuredSecret)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *configOidcResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.ConfigOidcModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
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
func (r *configOidcResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	httpResponse, err := r.Client.ConfigOIDCAPI.DeleteOidcConfiguration(r.AuthContext(ctx)).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error deleting OIDC configuration",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}
	if httpResponse != nil && httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Unexpected response when deleting OIDC configuration",
			fmt.Sprintf("Expected status 204, got %d", httpResponse.StatusCode),
		)
		return
	}
}

func (r *configOidcResource) doRead(ctx context.Context, respState *tfsdk.State, respDiags *diag.Diagnostics) *sonatypeiq.SsoConfigurationDTO {
	apiResponse, httpResponse, err := r.Client.ConfigOIDCAPI.GetOidcConfiguration(r.AuthContext(ctx)).Execute()

	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			respState.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"OIDC configuration did not exist",
				&err,
				httpResponse,
				respDiags,
			)
		} else {
			errors.HandleAPIError(
				"Error reading OIDC configuration",
				&err,
				httpResponse,
				respDiags,
			)
		}
		return nil
	}

	return apiResponse
}

func (r *configOidcResource) doUpsert(ctx context.Context, model *model.ConfigOidcModel, respState *tfsdk.State, respDiags *diag.Diagnostics) {
	httpResponse, err := r.Client.ConfigOIDCAPI.InsertOrUpdateOidcConfiguration(r.AuthContext(ctx)).
		SsoConfigurationDTO(*model.MapToApi()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating/updating OIDC configuration",
			&err,
			httpResponse,
			respDiags,
		)
		return
	} else if httpResponse.StatusCode != http.StatusNoContent {
		errors.HandleAPIError(
			"Upsertion of OIDC configuration was not successful",
			&err,
			httpResponse,
			respDiags,
		)
		return
	}

	// Preserve the configured secret - server returns "*******" mask
	configuredSecret := model.ClientSecret.ValueString()

	apiResponse := r.doRead(ctx, respState, respDiags)
	if apiResponse == nil {
		return
	}

	// Map response to State
	model.MapFromApi(apiResponse, configuredSecret)
	model.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}

func (r *configOidcResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
