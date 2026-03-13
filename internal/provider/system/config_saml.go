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
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// securitySamlResource is the resource implementation.
type securitySamlResource struct {
	common.BaseResource
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
			"id": sharedrschema.ResourceComputedString("Internal ID for Terraform State"),
			"identity_provider_name": sharedrschema.ResourceRequiredStringWithLengthAtLeast(
				"Hostname of the Proxy Server", 2,
			),
			"idp_metadata": sharedrschema.ResourceRequiredStringWithValidators(
				"SAML Identity Provider Metadata XML",
				stringvalidator.LengthAtLeast(10),
				stringvalidator.RegexMatches(
					regexp.MustCompile(`(?s)^\s*<.*>\s*$`),
					"must be valid XML format",
				),
			),
			"username_attribute": sharedrschema.ResourceRequiredStringWithLengthAtLeast(
				"IdP field mappings for username", 1,
			),
			"first_name_attribute": func() schema.StringAttribute {
				attr := sharedrschema.ResourceComputedOptionalString("IdP field mappings for user's given name")
				attr.Default = stringdefault.StaticString(common.SAML_DEFAULT_FIRST_NAME_ATTRIBUTE)
				attr.Validators = []validator.String{stringvalidator.LengthAtLeast(1)}
				return attr
			}(),
			"last_name_attribute": func() schema.StringAttribute {
				attr := sharedrschema.ResourceComputedOptionalString("IdP field mappings for user's family name")
				attr.Default = stringdefault.StaticString(common.SAML_DEFAULT_LAST_NAME_ATTRIBUTE)
				attr.Validators = []validator.String{stringvalidator.LengthAtLeast(1)}
				return attr
			}(),
			"email_attribute": func() schema.StringAttribute {
				attr := sharedrschema.ResourceComputedOptionalString("IdP field mappings for user's email")
				attr.Default = stringdefault.StaticString(common.SAML_DEFAULT_EMAIL_ATTRIBUTE)
				attr.Validators = []validator.String{stringvalidator.LengthAtLeast(1)}
				return attr
			}(),
			"groups_attribute": func() schema.StringAttribute {
				attr := sharedrschema.ResourceComputedOptionalString("IdP field mappings for user's groups")
				attr.Default = stringdefault.StaticString(common.SAML_DEFAULT_GROUPS_ATTRIBUTE)
				attr.Validators = []validator.String{stringvalidator.LengthAtLeast(1)}
				return attr
			}(),
			"validate_response_signature": sharedrschema.ResourceOptionalBoolWithDefault(
				"Validate SAML response signature", false,
			),
			"validate_assertion_signature": sharedrschema.ResourceOptionalBoolWithDefault(
				"By default, if a signing key is found in the IdP metadata, then Sonatype Nexus Repository Manager will attempt to validate signatures on the assertions.",
				false,
			),
			"entity_id": sharedrschema.ResourceRequiredStringWithLengthAtLeast(
				"SAML Entity ID (typically a URI)",
				1,
			),
			"last_updated": sharedrschema.ResourceLastUpdated(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *securitySamlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.ConfigSamlModel
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
func (r *securitySamlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.ConfigSamlModel
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
func (r *securitySamlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.ConfigSamlModel
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
func (r *securitySamlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	httpResponse, err := r.Client.ConfigSAMLAPI.DeleteSamlConfiguration(r.AuthContext(ctx)).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			common.ERR_SAML_CONFIGURATION_DID_NOT_EXIST,
			fmt.Sprintf("%v", err),
		)
		return
	}
}

func (r *securitySamlResource) doRead(ctx context.Context, respState *tfsdk.State, respDiags *diag.Diagnostics) *sonatypeiq.ApiSamlConfigurationResponseDTO {
	apiResponse, httpResponse, err := r.Client.ConfigSAMLAPI.GetSamlConfiguration(r.AuthContext(ctx)).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			respState.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"SAML configuration did not exist",
				&err,
				httpResponse,
				respDiags,
			)
		} else {
			errors.HandleAPIError(
				common.ERR_SAML_CONFIGURATION_DID_NOT_EXIST,
				&err,
				httpResponse,
				respDiags,
			)
		}
		return nil
	}

	return apiResponse
}

func (r *securitySamlResource) doUpsert(ctx context.Context, model *model.ConfigSamlModel, respState *tfsdk.State, respDiags *diag.Diagnostics) {
	httpResponse, err := r.Client.ConfigSAMLAPI.InsertOrUpdateSamlConfiguration(r.AuthContext(ctx)).IdentityProviderXml(model.IdpMetadata.ValueString()).SamlConfiguration(*model.MapToApi()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating/updating SAML configuration",
			&err,
			httpResponse,
			respDiags,
		)
		return
	} else if httpResponse.StatusCode != http.StatusNoContent {
		errors.HandleAPIError(
			"Upsertion of SAML configuration was not successful",
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

func (r *securitySamlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
