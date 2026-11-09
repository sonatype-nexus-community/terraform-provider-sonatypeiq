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

package model

import (
	"terraform-provider-sonatypeiq/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// ConfigOidcModel maps the OIDC configuration resource schema.
type ConfigOidcModel struct {
	ID                         types.String `tfsdk:"id"`
	ClientId                   types.String `tfsdk:"client_id"`
	ClientSecret               types.String `tfsdk:"client_secret"`
	IdpIssuer                  types.String `tfsdk:"idp_issuer"`
	IdpAuthorizationUrl        types.String `tfsdk:"idp_authorization_url"`
	IdpTokenUrl                types.String `tfsdk:"idp_token_url"`
	AuthorizationCustomParams  types.String `tfsdk:"authorization_custom_params_json"`
	TokenRequestCustomParams   types.String `tfsdk:"token_request_custom_params_json"`
	Oauth2IdpJwksUrl           types.String `tfsdk:"oauth2_idp_jwks_url"`
	Oauth2IdpJwks              types.String `tfsdk:"oauth2_idp_jwks"`
	Oauth2IdpJwsAlgorithm      types.String `tfsdk:"oauth2_idp_jws_algorithm"`
	Oauth2ExpectedAudience     types.String `tfsdk:"oauth2_expected_audience"`
	Oauth2UsernameClaim        types.String `tfsdk:"oauth2_username_claim"`
	Oauth2EmailClaim           types.String `tfsdk:"oauth2_email_claim"`
	Oauth2FirstNameClaim       types.String `tfsdk:"oauth2_first_name_claim"`
	Oauth2LastNameClaim        types.String `tfsdk:"oauth2_last_name_claim"`
	Oauth2GroupsClaim          types.String `tfsdk:"oauth2_groups_claim"`
	Oauth2ExactMatchClaimsJson types.String `tfsdk:"oauth2_exact_match_claims_json"`
	LastUpdated                types.String `tfsdk:"last_updated"`
}

// setIfPresent passes the value to set only when it is configured.
func setIfPresent(value types.String, set func(string)) {
	if !value.IsNull() {
		set(value.ValueString())
	}
}

// MapToApi converts the Terraform model to the SsoConfigurationDTO for the API.
func (m *ConfigOidcModel) MapToApi() *sonatypeiq.SsoConfigurationDTO {
	ssoConfig := sonatypeiq.NewSsoConfigurationDTOWithDefaults()

	// Build OIDC configuration
	oidcConfig := sonatypeiq.NewOidcConfigurationDTOWithDefaults()
	setIfPresent(m.ClientId, oidcConfig.SetClientId)
	setIfPresent(m.ClientSecret, oidcConfig.SetClientSecret)
	setIfPresent(m.IdpIssuer, oidcConfig.SetIdpIssuer)
	setIfPresent(m.IdpAuthorizationUrl, oidcConfig.SetIdpAuthorizationUrl)
	setIfPresent(m.IdpTokenUrl, oidcConfig.SetIdpTokenUrl)
	setIfPresent(m.AuthorizationCustomParams, oidcConfig.SetAuthorizationCustomParamsJson)
	setIfPresent(m.TokenRequestCustomParams, oidcConfig.SetTokenRequestCustomParamsJson)
	ssoConfig.SetOidcConfiguration(*oidcConfig)

	// Build OAuth2 configuration
	oauth2Config := sonatypeiq.NewOAuth2ConfigurationDTOWithDefaults()
	setIfPresent(m.Oauth2IdpJwksUrl, oauth2Config.SetIdpJwksUrl)
	setIfPresent(m.Oauth2IdpJwks, oauth2Config.SetIdpJwks)
	setIfPresent(m.Oauth2IdpJwsAlgorithm, oauth2Config.SetIdpJwsAlgorithm)
	setIfPresent(m.Oauth2ExpectedAudience, oauth2Config.SetExpectedAudience)
	setIfPresent(m.Oauth2UsernameClaim, oauth2Config.SetUsernameClaim)
	setIfPresent(m.Oauth2EmailClaim, oauth2Config.SetEmailClaim)
	setIfPresent(m.Oauth2FirstNameClaim, oauth2Config.SetFirstNameClaim)
	setIfPresent(m.Oauth2LastNameClaim, oauth2Config.SetLastNameClaim)
	setIfPresent(m.Oauth2GroupsClaim, oauth2Config.SetGroupsClaim)
	setIfPresent(m.Oauth2ExactMatchClaimsJson, oauth2Config.SetExactMatchClaimsJson)
	// idp_issuer is present in both oidc and oauth2 configurations; prefer the oidc one
	setIfPresent(m.IdpIssuer, oauth2Config.SetIdpIssuer)
	ssoConfig.SetOauth2Configuration(*oauth2Config)

	return ssoConfig
}

// MapFromApi populates the Terraform model from the API response.
// Note: client_secret is never returned by the server (it returns "*******" as a mask),
// so we preserve the configured value in state.
func (m *ConfigOidcModel) MapFromApi(api *sonatypeiq.SsoConfigurationDTO, configuredSecret string) {
	m.ID = types.StringValue(common.STATE_ID_OIDC_CONFIGURATION)

	// Map OIDC configuration
	if oidc, ok := api.GetOidcConfigurationOk(); ok && oidc != nil {
		m.ClientId = types.StringPointerValue(oidc.ClientId)
		// Preserve the configured secret - server returns "*******" mask
		m.ClientSecret = types.StringValue(configuredSecret)
		m.IdpIssuer = types.StringPointerValue(oidc.IdpIssuer)
		m.IdpAuthorizationUrl = types.StringPointerValue(oidc.IdpAuthorizationUrl)
		m.IdpTokenUrl = types.StringPointerValue(oidc.IdpTokenUrl)
		m.AuthorizationCustomParams = types.StringPointerValue(oidc.AuthorizationCustomParamsJson)
		m.TokenRequestCustomParams = types.StringPointerValue(oidc.TokenRequestCustomParamsJson)
	}

	// Map OAuth2 configuration
	if oauth2, ok := api.GetOauth2ConfigurationOk(); ok && oauth2 != nil {
		m.Oauth2IdpJwksUrl = types.StringPointerValue(oauth2.IdpJwksUrl)
		m.Oauth2IdpJwks = types.StringPointerValue(oauth2.IdpJwks)
		m.Oauth2IdpJwsAlgorithm = types.StringPointerValue(oauth2.IdpJwsAlgorithm)
		m.Oauth2ExpectedAudience = types.StringPointerValue(oauth2.ExpectedAudience)
		m.Oauth2UsernameClaim = types.StringPointerValue(oauth2.UsernameClaim)
		m.Oauth2EmailClaim = types.StringPointerValue(oauth2.EmailClaim)
		m.Oauth2FirstNameClaim = types.StringPointerValue(oauth2.FirstNameClaim)
		m.Oauth2LastNameClaim = types.StringPointerValue(oauth2.LastNameClaim)
		m.Oauth2GroupsClaim = types.StringPointerValue(oauth2.GroupsClaim)
		m.Oauth2ExactMatchClaimsJson = types.StringPointerValue(oauth2.ExactMatchClaimsJson)
	}
}
