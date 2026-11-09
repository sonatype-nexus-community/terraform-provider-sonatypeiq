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

package system_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
)

// fakeOidcIqServer is a minimal fake IQ Server for OIDC configuration testing.
// The outer HTTP handler holds the mutex, so inner handlers must NOT lock it again.
type fakeOidcIqServer struct {
	*httptest.Server
	mu           sync.Mutex
	config       *sonatypeiq.SsoConfigurationDTO
	storedConfig *sonatypeiq.SsoConfigurationDTO // what was stored (before masking)
	deleted      bool
	release      int
}

func newFakeOidcIqServer(t *testing.T, release int) *fakeOidcIqServer {
	t.Helper()
	f := &fakeOidcIqServer{release: release}

	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", fmt.Sprintf("NexusIQ/1.%d.0-01", release))
		f.mu.Lock()
		defer f.mu.Unlock()
		f.route(w, r)
	}))
	t.Cleanup(f.Close)

	t.Setenv("IQ_SERVER_URL", f.URL)
	t.Setenv("IQ_SERVER_USERNAME", "admin")
	t.Setenv("IQ_SERVER_PASSWORD", "admin123")
	return f
}

// route dispatches requests. Caller holds f.mu.
func (f *fakeOidcIqServer) route(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/ping":
		return // anonymous
	case "/api/v2/userTokens/currentUser/hasToken":
		http.Error(w, "no token", http.StatusNotFound) // credentials check
		return
	case "/api/v2/users/admin":
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"username":"admin"}`))
		return
	case "/api/v2/config/oidc":
		f.handleOidcConfig(w, r)
		return
	default:
		http.NotFound(w, r)
	}
}

// handleOidcConfig dispatches OIDC config requests. Caller holds f.mu.
func (f *fakeOidcIqServer) handleOidcConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		f.handleGet(w, r)
	case http.MethodPut:
		f.handlePut(w, r)
	case http.MethodDelete:
		f.handleDelete(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleGet handles GET requests. Caller holds f.mu.
func (f *fakeOidcIqServer) handleGet(w http.ResponseWriter, r *http.Request) {
	if f.config == nil {
		http.Error(w, "Oidc configuration not set", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(f.config)
}

// handlePut handles PUT requests. Caller holds f.mu.
func (f *fakeOidcIqServer) handlePut(w http.ResponseWriter, r *http.Request) {
	var config sonatypeiq.SsoConfigurationDTO
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if msg := validateFakeOidcConfig(&config); msg != "" {
		http.Error(w, "Invalid OIDC configuration: "+msg, http.StatusBadRequest)
		return
	}

	// Store the original config (for testing what was sent)
	f.storedConfig = &config

	// Copy config and mask the client secret like the real server does
	responseConfig := config
	if responseConfig.OidcConfiguration != nil && responseConfig.OidcConfiguration.ClientSecret != nil {
		masked := "*******"
		responseConfig.OidcConfiguration.ClientSecret = &masked
	}

	f.config = &responseConfig
	f.deleted = false
	w.WriteHeader(http.StatusNoContent)
}

// handleDelete handles DELETE requests. Caller holds f.mu.
func (f *fakeOidcIqServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	if f.config == nil {
		http.Error(w, "Oidc configuration not set", http.StatusNotFound)
		return
	}

	f.deleted = true
	f.config = nil
	f.storedConfig = nil
	w.WriteHeader(http.StatusNoContent)
}

// storedValue returns the value of a field in the configuration the fake server stored, or "" when it is not set.
func storedValue(config *sonatypeiq.SsoConfigurationDTO, field string) (string, bool) {
	oidc := config.GetOidcConfiguration()
	oauth2 := config.GetOauth2Configuration()
	switch field {
	case "client_id":
		return oidc.GetClientId(), true
	case "idp_issuer":
		return oidc.GetIdpIssuer(), true
	case "authorization_custom_params_json":
		return oidc.GetAuthorizationCustomParamsJson(), true
	case "oauth2_username_claim":
		return oauth2.GetUsernameClaim(), true
	}
	return "", false
}

func (f *fakeOidcIqServer) checkStored(field, want string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.storedConfig == nil {
			return fmt.Errorf("no OIDC configuration stored")
		}
		got, known := storedValue(f.storedConfig, field)
		if !known {
			return fmt.Errorf("unknown field: %s", field)
		}
		if got != want {
			return fmt.Errorf("server has %s = %q, want %q", field, got, want)
		}
		return nil
	}
}

const fakeOidcConfigMinimal = `
resource "sonatypeiq_config_oidc" "test" {
  client_id             = "test-client-id"
  client_secret         = "test-client-secret"
  idp_issuer            = "https://idp.example.com"
  idp_authorization_url = "https://idp.example.com/authorize"
  idp_token_url         = "https://idp.example.com/token"
  oauth2_idp_jws_algorithm = "RS256"
  oauth2_idp_jwks_url = "https://idp.example.com/jwks"
}
`

const fakeOidcConfigWithOAuth2 = `
resource "sonatypeiq_config_oidc" "test" {
  client_id               = "test-client-id"
  client_secret           = "test-client-secret"
  idp_issuer              = "https://idp.example.com"
  idp_authorization_url   = "https://idp.example.com/authorize"
  idp_token_url           = "https://idp.example.com/token"
  oauth2_idp_jws_algorithm = "RS256"
  oauth2_idp_jwks_url     = "https://idp.example.com/jwks"
  oauth2_username_claim   = "preferred_username"
  oauth2_email_claim      = "email"
}
`

const fakeOidcConfigWithCustomParams = `
resource "sonatypeiq_config_oidc" "test" {
  client_id                          = "test-client-id"
  client_secret                      = "test-client-secret"
  idp_issuer                         = "https://idp.example.com"
  idp_authorization_url              = "https://idp.example.com/authorize"
  idp_token_url                      = "https://idp.example.com/token"
  oauth2_idp_jws_algorithm = "RS256"
  oauth2_idp_jwks_url = "https://idp.example.com/jwks"
  authorization_custom_params_json   = "{\"prompt\":\"login\"}"
  token_request_custom_params_json   = "{\"audience\":\"iq-server\"}"
}
`

const fakeOidcConfigUpdated = `
resource "sonatypeiq_config_oidc" "test" {
  client_id             = "updated-client-id"
  client_secret         = "updated-client-secret"
  idp_issuer            = "https://idp.example.com"
  idp_authorization_url = "https://idp.example.com/authorize"
  idp_token_url         = "https://idp.example.com/token"
  oauth2_idp_jws_algorithm = "RS256"
  oauth2_idp_jwks_url = "https://idp.example.com/jwks"
}
`

// OIDC is only supported from release 198 onwards - gating works at plan time.
func TestConfigOidcResource_RequiresMinimumVersion_Offline(t *testing.T) {
	newFakeOidcIqServer(t, 190)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      fakeOidcConfigMinimal,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`is not supported by this Sonatype IQ Server version`),
			},
		},
	})
}

// On an unknown server version (release 0), the resource is NOT blocked.
func TestConfigOidcResource_NotBlockedOnUnknownVersion_Offline(t *testing.T) {
	newFakeOidcIqServer(t, 0) // version 0 = unknown

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fakeOidcConfigMinimal,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sonatypeiq_config_oidc.test", "client_id", "test-client-id"),
				),
			},
		},
	})
}

// OIDC configuration can be created and deleted on a supported server version.
func TestConfigOidcResource_CreateAndDelete_Offline(t *testing.T) {
	f := newFakeOidcIqServer(t, 205)
	resourceName := "sonatypeiq_config_oidc.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fakeOidcConfigMinimal,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_id", "test-client-id"),
					resource.TestCheckResourceAttr(resourceName, "client_secret", "test-client-secret"),
					resource.TestCheckResourceAttr(resourceName, "idp_issuer", "https://idp.example.com"),
					f.checkStored("client_id", "test-client-id"),
				),
			},
		},
	})
}

// OIDC configuration can be updated.
func TestConfigOidcResource_Update_Offline(t *testing.T) {
	f := newFakeOidcIqServer(t, 205)
	resourceName := "sonatypeiq_config_oidc.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fakeOidcConfigMinimal,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_id", "test-client-id"),
				),
			},
			{
				Config: fakeOidcConfigUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_id", "updated-client-id"),
					resource.TestCheckResourceAttr(resourceName, "client_secret", "updated-client-secret"),
					f.checkStored("client_id", "updated-client-id"),
				),
			},
		},
	})
}

// OAuth2 fields can be configured and are stored correctly.
func TestConfigOidcResource_WithOAuth2Fields_Offline(t *testing.T) {
	f := newFakeOidcIqServer(t, 205)
	resourceName := "sonatypeiq_config_oidc.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fakeOidcConfigWithOAuth2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_id", "test-client-id"),
					resource.TestCheckResourceAttr(resourceName, "oauth2_idp_jwks_url", "https://idp.example.com/jwks"),
					resource.TestCheckResourceAttr(resourceName, "oauth2_username_claim", "preferred_username"),
					f.checkStored("oauth2_username_claim", "preferred_username"),
				),
			},
		},
	})
}

// Custom JSON param fields are stored correctly and not rewritten.
func TestConfigOidcResource_CustomJsonFields_Offline(t *testing.T) {
	f := newFakeOidcIqServer(t, 205)
	resourceName := "sonatypeiq_config_oidc.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fakeOidcConfigWithCustomParams,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "authorization_custom_params_json", "{\"prompt\":\"login\"}"),
					resource.TestCheckResourceAttr(resourceName, "token_request_custom_params_json", "{\"audience\":\"iq-server\"}"),
					f.checkStored("authorization_custom_params_json", "{\"prompt\":\"login\"}"),
				),
			},
			// No drift - JSON fields are preserved
			{
				Config:   fakeOidcConfigWithCustomParams,
				PlanOnly: true,
			},
		},
	})
}

// Client secret is preserved from config, not replaced by the server's mask.
func TestConfigOidcResource_ClientSecretPreserved_Offline(t *testing.T) {
	newFakeOidcIqServer(t, 205)
	resourceName := "sonatypeiq_config_oidc.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fakeOidcConfigMinimal,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_secret", "test-client-secret"),
				),
			},
			// Re-apply - secret must still be in state (not replaced by mask)
			{
				Config: fakeOidcConfigMinimal,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_secret", "test-client-secret"),
				),
			},
		},
	})
}

// Optional attributes can be removed on update (PUT replaces the configuration).
func TestConfigOidcResource_RemoveOptionalField_Offline(t *testing.T) {
	f := newFakeOidcIqServer(t, 205)
	resourceName := "sonatypeiq_config_oidc.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with an optional field
			{
				Config: fakeOidcConfigWithOAuth2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_id", "test-client-id"),
					resource.TestCheckResourceAttr(resourceName, "oauth2_username_claim", "preferred_username"),
					f.checkStored("oauth2_username_claim", "preferred_username"),
				),
			},
			// Update to remove the optional field
			{
				Config: fakeOidcConfigMinimal,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_id", "test-client-id"),
					resource.TestCheckNoResourceAttr(resourceName, "oauth2_username_claim"),
				),
			},
		},
	})
}

// Import works, but client_secret stays null after import (it cannot be read back).
func TestConfigOidcResource_Import_Offline(t *testing.T) {
	_ = newFakeOidcIqServer(t, 205)
	resourceName := "sonatypeiq_config_oidc.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// First create
			{
				Config: fakeOidcConfigMinimal,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_id", "test-client-id"),
				),
			},
			// Import - client_secret will be unknown/null after import since server doesn't return it
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated", "client_secret"},
			},
		},
	})
}

// validateFakeOidcConfig applies the checks the real server makes before it stores a configuration.
func validateFakeOidcConfig(config *sonatypeiq.SsoConfigurationDTO) string {
	oauth2 := config.Oauth2Configuration
	switch {
	case oauth2 == nil || oauth2.IdpJwsAlgorithm == nil || *oauth2.IdpJwsAlgorithm == "":
		return "The IDP JWS Algorithm is required"
	case strings.EqualFold(*oauth2.IdpJwsAlgorithm, "none"):
		return "JWS Algorithm not allowed"
	case (oauth2.IdpJwksUrl == nil || *oauth2.IdpJwksUrl == "") && (oauth2.IdpJwks == nil || *oauth2.IdpJwks == ""):
		return "Either the IDP JWKS Url or a JWKS string should be provided"
	}
	return ""
}

// The server rejects a configuration with no JWS algorithm, with the algorithm "none", or with neither a JWKS URL
// nor a JWKS, so each is refused at plan time.
func TestConfigOidcResource_RejectsWhatTheServerRejects_Offline(t *testing.T) {
	newFakeOidcIqServer(t, 209)

	const base = `
resource "sonatypeiq_config_oidc" "test" {
  client_id             = "test-client-id"
  client_secret         = "test-client-secret"
  idp_issuer            = "https://idp.example.com"
  idp_authorization_url = "https://idp.example.com/authorize"
  idp_token_url         = "https://idp.example.com/token"
`
	tests := map[string]struct {
		extra string
		error string
	}{
		"no algorithm":     {`  oauth2_idp_jwks_url = "https://idp.example.com/jwks"` + "\n}", `(?s)oauth2_idp_jws_algorithm.*is required`},
		"algorithm none":   {`  oauth2_idp_jws_algorithm = "None"` + "\n" + `  oauth2_idp_jwks_url = "https://idp.example.com/jwks"` + "\n}", `(?s)oauth2_idp_jws_algorithm`},
		"no jwks anywhere": {`  oauth2_idp_jws_algorithm = "RS256"` + "\n}", `(?s)oauth2_idp_jwks_url.*oauth2_idp_jwks`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: defaultProvider + base + tt.extra, ExpectError: regexp.MustCompile(tt.error)},
				},
			})
		})
	}
}
