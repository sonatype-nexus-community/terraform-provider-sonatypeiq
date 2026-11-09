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
	"fmt"
	"net/http"
	"os"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/testutil"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccConfigOidcResource tests the OIDC configuration resource against a real IQ Server.
// SAFETY: This test uses dummy, non-working OIDC values that will not affect authentication.
// It first checks that OIDC is not already configured, and cleans up after itself.
// The test verifies that authentication still works after configuration.
func TestAccConfigOidcResource(t *testing.T) {
	// Keep in step with oidcIntroducedIn in the resource
	testutil.SkipIfNxiqVersionOlderThan(t, 198)

	// Safety check: ensure OIDC is not already configured
	serverUrl := os.Getenv("IQ_SERVER_URL")
	username := os.Getenv("IQ_SERVER_USERNAME")
	password := os.Getenv("IQ_SERVER_PASSWORD")
	if serverUrl == "" || username == "" || password == "" {
		t.Skip("IQ_SERVER_URL, IQ_SERVER_USERNAME and IQ_SERVER_PASSWORD must be set")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", serverUrl+"/api/v2/config/oidc", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to check OIDC configuration: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// If OIDC is already configured (200 OK), skip the test to avoid affecting a live configuration
	if resp.StatusCode == http.StatusOK {
		t.Skip("OIDC is already configured on this server - skipping test to avoid affecting live configuration")
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Unexpected status %d when checking OIDC configuration", resp.StatusCode)
	}

	resourceName := "sonatypeiq_config_oidc.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with minimal configuration (non-functional dummy values)
			{
				Config: testAccConfigOidcMinimalResource(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_OIDC_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "client_id", "dummy-client-id"),
					resource.TestCheckResourceAttr(resourceName, "idp_issuer", "https://dummy.example.com"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
					// Verify authentication still works after OIDC configuration
					verifyAuthenticationWorks(serverUrl, username, password),
				),
			},
			// Import testing
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_secret", "last_updated"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})

	// Final safety check: ensure OIDC is cleaned up
	req2, err := http.NewRequest("DELETE", serverUrl+"/api/v2/config/oidc", nil)
	if err == nil {
		req2.SetBasicAuth(username, password)
		resp2, err := client.Do(req2)
		if err == nil && resp2.StatusCode == http.StatusNoContent {
			t.Logf("UTC %s: Cleaned up OIDC configuration after test", time.Now().UTC().Format(time.RFC3339))
		}
		if resp2 != nil {
			_ = resp2.Body.Close()
		}
	}
}

func testAccConfigOidcMinimalResource() string {
	return utils_test.ProviderConfig + `
resource "sonatypeiq_config_oidc" "test" {
  client_id             = "dummy-client-id"
  client_secret         = "dummy-client-secret"
  idp_issuer            = "https://dummy.example.com"
  idp_authorization_url = "https://dummy.example.com/authorize"
  idp_token_url         = "https://dummy.example.com/token"
  oauth2_idp_jws_algorithm = "RS256"
  oauth2_idp_jwks_url = "https://idp.example.com/jwks"
}`
}

// verifyAuthenticationWorks checks that the IQ Server still accepts credentials after OIDC configuration.
// It calls /api/v2/userTokens/currentUser/hasToken which should return 200 or 404 (not 401).
func verifyAuthenticationWorks(serverUrl, username, password string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequest("GET", serverUrl+"/api/v2/userTokens/currentUser/hasToken", nil)
		if err != nil {
			return fmt.Errorf("failed to create auth check request: %v", err)
		}
		req.SetBasicAuth(username, password)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to verify authentication: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// 200 or 404 with a body = authentication works
		// 401 = authentication broken (OIDC took over incorrectly)
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("authentication broken after OIDC configuration (got 401 from /userTokens/currentUser/hasToken) - OIDC may have affected sign-in")
		}

		// 200 or 404 are both acceptable
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("unexpected status %d from auth check endpoint", resp.StatusCode)
		}

		return nil
	}
}
