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
	"regexp"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// securitySamlResource is the resource implementation.
type securitySamlResource struct {
	common.BaseResourceWithImport
}

// NewConfigSamlResource is a helper function to simplify the provider implementation.
func NewConfigSamlResource() resource.Resource {
	return &securitySamlResource{}
}

// Metadata returns the resource type name.
func (r *securitySamlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_saml"
}

// Schema defines the schema for the resource.
func (r *securitySamlResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configure Sonatype IQ SAML connection.",
		Attributes: map[string]schema.Attribute{
			"identity_provider_name": schema.StringAttribute{
				Description: "The name of the Identity Provider that is displayed on the login page when SAML is configured",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(2),
				},
			},
			"idp_metadata": schema.StringAttribute{
				Description: "SAML Identity Provider Metadata XML",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(10),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`(?s)^\s*<.*>\s*$`),
						"must be valid XML format",
					),
				},
			},
			"username_attribute": schema.StringAttribute{
				Description: "IdP field mappings for username",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"first_name_attribute": schema.StringAttribute{
				Description: "IdP field mappings for user's given name",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"last_name_attribute": schema.StringAttribute{
				Description: "IdP field mappings for user's family name",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"email_attribute": schema.StringAttribute{
				Description: "IdP field mappings for user's email",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"groups_attribute": schema.StringAttribute{
				Description: "IdP field mappings for user's groups",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"validate_response_signature": schema.BoolAttribute{
				Description: "Validate SAML response signature",
				Optional:    true,
			},
			"validate_assertion_signature": schema.BoolAttribute{
				Description: "By default, if a signing key is found in the IdP metadata, then Sonatype Nexus Repository Manager will attempt to validate signatures on the assertions.",
				Optional:    true,
			},
			"entity_id": schema.StringAttribute{
				Description: "SAML Entity ID (typically a URI)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *securitySamlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.SecuritySamlModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	r.upsert(ctx, plan, &resp.Diagnostics, &resp.State)
}

// Read refreshes the Terraform state with the latest data.
func (r *securitySamlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.read(ctx, &resp.Diagnostics, &resp.State)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *securitySamlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.SecuritySamlModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	r.upsert(ctx, plan, &resp.Diagnostics, &resp.State)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *securitySamlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	httpResponse, err := r.Client.ConfigSAMLAPI.DeleteSamlConfiguration(ctx).Execute()
	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		common.HandleApiError(
			"Error deleting SAML configuration",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}
}

func (r *securitySamlResource) read(ctx context.Context, respDiags *diag.Diagnostics, respState *tfsdk.State) {
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	apiResponse, httpResponse, err := r.Client.ConfigSAMLAPI.GetSamlConfiguration(ctx).Execute()
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			respState.RemoveResource(ctx)
			common.HandleApiWarning(
				"SAML Configuration did not exist",
				&err,
				httpResponse,
				respDiags,
			)
		} else {
			common.HandleApiError(
				"Failed to read SAML configuration",
				&err,
				httpResponse,
				respDiags,
			)
		}
		return
	}

	state := model.SecuritySamlModel{}
	state.MapFromApi(apiResponse)
	respDiags.Append(respState.Set(ctx, &state)...)
}

func (r *securitySamlResource) upsert(ctx context.Context, plan model.SecuritySamlModel, respDiags *diag.Diagnostics, respState *tfsdk.State) {
	// Set up authentication context
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	apiSamlRequest := sonatypeiq.NewApiSamlConfigurationDTOWithDefaults()
	plan.MapToApi(apiSamlRequest)

	httpResponse, err := r.Client.ConfigSAMLAPI.InsertOrUpdateSamlConfiguration(ctx).IdentityProviderXml(plan.IdpMetadata.ValueString()).SamlConfiguration(*apiSamlRequest).Execute()
	if err != nil {
		common.HandleApiError(
			"Error creating SAML configuration",
			&err,
			httpResponse,
			respDiags,
		)
		return
	} else if httpResponse.StatusCode != http.StatusNoContent {
		common.HandleApiError(
			"Error creating SAML configuration - unexpected response code",
			&err,
			httpResponse,
			respDiags,
		)
		return
	}

	diags := respState.Set(ctx, plan)
	respDiags.Append(diags...)
}

func (r *securitySamlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.read(ctx, &resp.Diagnostics, &resp.State)
}
