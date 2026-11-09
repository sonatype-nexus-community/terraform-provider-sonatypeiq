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
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"terraform-provider-sonatypeiq/internal/provider/common"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// fakeReverseProxyAuthIqServer is just enough of Sonatype IQ Server to drive
// sonatypeiq_config_reverse_proxy_authentication end to end through the Terraform CLI
// without a licensed server.
type fakeReverseProxyAuthIqServer struct {
	*httptest.Server
	mu        sync.Mutex
	config    map[string]any
	requested bool
	deleted   bool
}

func newFakeReverseProxyAuthIqServer(t *testing.T) *fakeReverseProxyAuthIqServer {
	t.Helper()
	f := &fakeReverseProxyAuthIqServer{config: map[string]any{}}

	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "NexusIQ/1.209.0-01")
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

// route dispatches a request. The caller holds the lock.
func (f *fakeReverseProxyAuthIqServer) route(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/ping":
		return // anonymous
	case r.URL.Path == "/api/v2/userTokens/currentUser/hasToken":
		http.Error(w, "no token", http.StatusNotFound) // the credentials check: reachable and accepted
		return
	case strings.HasPrefix(r.URL.Path, "/api/v2/users/"):
		f.handleUser(w)
		return
	case r.URL.Path == "/api/v2/config/reverseProxyAuthentication":
		switch r.Method {
		case http.MethodPut:
			f.handlePut(w, r)
		case http.MethodGet:
			f.handleGet(w, r)
		case http.MethodDelete:
			f.handleDelete(w, r)
		default:
			http.NotFound(w, r)
		}
		return
	}
	http.NotFound(w, r)
}

func (f *fakeReverseProxyAuthIqServer) handleUser(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"username":"admin"}`))
}

func (f *fakeReverseProxyAuthIqServer) handlePut(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Real server behavior: PUT replaces enabled, csrfProtectionDisabled, and logoutUrl
	// on every request, but only changes usernameHeader when one is sent.
	// When usernameHeader is absent from the PUT, the stored value is kept.
	if _, ok := body["usernameHeader"]; !ok {
		// Preserve existing usernameHeader if present
		if existingHeader, exists := f.config["usernameHeader"]; exists {
			body["usernameHeader"] = existingHeader
		}
	}
	// Always replace the other three fields
	f.config = body
	f.requested = true
	w.WriteHeader(http.StatusNoContent)
}

func (f *fakeReverseProxyAuthIqServer) handleGet(w http.ResponseWriter, r *http.Request) {
	if !f.requested {
		http.Error(w, "Reverse proxy authentication not configured.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	body := map[string]any{
		"enabled":                f.config["enabled"],
		"usernameHeader":         f.config["usernameHeader"],
		"csrfProtectionDisabled": f.config["csrfProtectionDisabled"],
		"logoutUrl":              f.config["logoutUrl"],
	}
	// Apply server-side default for usernameHeader
	if body["usernameHeader"] == nil {
		body["usernameHeader"] = "REMOTE_USER"
	}
	_ = json.NewEncoder(w).Encode(body)
}

func (f *fakeReverseProxyAuthIqServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	f.deleted = true
	// DELETE resets to server defaults (clears all stored values)
	f.config = map[string]any{}
	f.requested = false
	w.WriteHeader(http.StatusNoContent)
}

func TestAccConfigReverseProxyAuthenticationResourceOffline(t *testing.T) {
	_ = newFakeReverseProxyAuthIqServer(t)
	resourceName := "sonatypeiq_config_reverse_proxy_authentication.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing - with enabled = false for safety
			{
				Config: testAccConfigReverseProxyAuthResourceMinimal(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_REVERSE_PROXY_AUTHENTICATION_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "username_header"),
					resource.TestCheckNoResourceAttr(resourceName, "csrf_protection_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "logout_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Update testing
			{
				Config: testAccConfigReverseProxyAuthResourceFull(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_REVERSE_PROXY_AUTHENTICATION_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "username_header", "X-Forwarded-User"),
					resource.TestCheckResourceAttr(resourceName, "csrf_protection_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "logout_url", "https://sso.example.com/logout"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Import testing
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
		},
	})
}

// TestAccConfigReverseProxyAuthenticationResourceOffline_RemoveUsernameHeader tests that
// removing username_header from the config properly resets it to the server default.
func TestAccConfigReverseProxyAuthenticationResourceOffline_RemoveUsernameHeader(t *testing.T) {
	_ = newFakeReverseProxyAuthIqServer(t)
	resourceName := "sonatypeiq_config_reverse_proxy_authentication.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Set username_header to a custom value
			{
				Config: testAccConfigReverseProxyAuthResourceWithUsernameHeader(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "username_header", "X-Custom-User"),
				),
			},
			// Remove username_header - should reset to server default
			{
				Config: testAccConfigReverseProxyAuthResourceMinimal(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "username_header"),
				),
			},
			// Refresh state from server to verify it's actually reset (no silent drift)
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "username_header"),
				),
			},
			// Verify no drift - a follow-up plan should be empty
			{
				Config:             testAccConfigReverseProxyAuthResourceMinimal(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccConfigReverseProxyAuthenticationResourceOffline_RemoveLogoutUrl tests that
// removing logout_url from the config properly clears it on the server.
func TestAccConfigReverseProxyAuthenticationResourceOffline_RemoveLogoutUrl(t *testing.T) {
	_ = newFakeReverseProxyAuthIqServer(t)
	resourceName := "sonatypeiq_config_reverse_proxy_authentication.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Set logout_url
			{
				Config: testAccConfigReverseProxyAuthResourceWithLogoutUrl(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "logout_url", "https://sso.example.com/logout"),
				),
			},
			// Remove logout_url - should be cleared
			{
				Config: testAccConfigReverseProxyAuthResourceMinimal(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "logout_url"),
				),
			},
			// Verify no drift
			{
				Config:             testAccConfigReverseProxyAuthResourceMinimal(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccConfigReverseProxyAuthenticationResourceOffline_RemoveCsrfProtectionDisabled tests that
// removing csrf_protection_disabled from the config properly clears it on the server.
func TestAccConfigReverseProxyAuthenticationResourceOffline_RemoveCsrfProtectionDisabled(t *testing.T) {
	_ = newFakeReverseProxyAuthIqServer(t)
	resourceName := "sonatypeiq_config_reverse_proxy_authentication.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Set csrf_protection_disabled
			{
				Config: testAccConfigReverseProxyAuthResourceWithCsrfDisabled(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "csrf_protection_disabled", "true"),
				),
			},
			// Remove csrf_protection_disabled - should be cleared
			{
				Config: testAccConfigReverseProxyAuthResourceMinimal(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "csrf_protection_disabled"),
				),
			},
			// Verify no drift
			{
				Config:             testAccConfigReverseProxyAuthResourceMinimal(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccConfigReverseProxyAuthResourceMinimal() string {
	return utils_test.ProviderConfig + `
resource "sonatypeiq_config_reverse_proxy_authentication" "test" {
  enabled = false
}`
}

func testAccConfigReverseProxyAuthResourceFull() string {
	return utils_test.ProviderConfig + `
resource "sonatypeiq_config_reverse_proxy_authentication" "test" {
  enabled                  = false
  username_header          = "X-Forwarded-User"
  csrf_protection_disabled = true
  logout_url               = "https://sso.example.com/logout"
}`
}

func testAccConfigReverseProxyAuthResourceWithUsernameHeader() string {
	return utils_test.ProviderConfig + `
resource "sonatypeiq_config_reverse_proxy_authentication" "test" {
  enabled         = false
  username_header = "X-Custom-User"
}`
}

func testAccConfigReverseProxyAuthResourceWithLogoutUrl() string {
	return utils_test.ProviderConfig + `
resource "sonatypeiq_config_reverse_proxy_authentication" "test" {
  enabled    = false
  logout_url = "https://sso.example.com/logout"
}`
}

func testAccConfigReverseProxyAuthResourceWithCsrfDisabled() string {
	return utils_test.ProviderConfig + `
resource "sonatypeiq_config_reverse_proxy_authentication" "test" {
  enabled                  = false
  csrf_protection_disabled = true
}`
}
