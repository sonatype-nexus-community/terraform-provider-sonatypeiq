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
package user_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"sync/atomic"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// unprivilegedIq is a Sonatype IQ Server for a user that has no administrator permissions: reading users is
// forbidden, but the current user's own token endpoints (which need no permission) work. This is the situation
// reported in #91.
type unprivilegedIq struct {
	*httptest.Server
	mu          sync.Mutex
	hasToken    bool
	usersCalls  atomic.Int32
	rejectLogin bool
}

func newUnprivilegedIq(t *testing.T, serverHeader string, rejectLogin bool) *unprivilegedIq {
	t.Helper()
	f := &unprivilegedIq{rejectLogin: rejectLogin}

	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", serverHeader)
		f.mu.Lock()
		defer f.mu.Unlock()
		f.route(w, r)
	}))
	t.Cleanup(f.Close)

	t.Setenv("IQ_SERVER_URL", f.URL)
	t.Setenv("IQ_SERVER_USERNAME", "limited")
	t.Setenv("IQ_SERVER_PASSWORD", "secret")
	return f
}

func (f *unprivilegedIq) route(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/ping" {
		return // anonymous
	}
	if f.rejectLogin {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch {
	case r.URL.Path == "/api/v2/userTokens/currentUser/hasToken":
		f.handleHasToken(w)
	case r.URL.Path == "/api/v2/userTokens/currentUser" && r.Method == http.MethodPost:
		f.hasToken = true
		writeJSON(w, map[string]any{"userCode": "abc123", "passCode": "def456", "username": "limited", "realm": "Internal"})
	case r.URL.Path == "/api/v2/userTokens/currentUser" && r.Method == http.MethodDelete:
		f.hasToken = false
		w.WriteHeader(http.StatusNoContent)
	default:
		// Everything else - including reading users - needs an administrator
		f.usersCalls.Add(1)
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
	}
}

func (f *unprivilegedIq) handleHasToken(w http.ResponseWriter) {
	if !f.hasToken {
		http.Error(w, "no token", http.StatusNotFound)
		return
	}
	writeJSON(w, map[string]any{"userTokenExists": true})
}

func writeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func userTokenConfig(providerBlock string) string {
	return providerBlock + `
resource "sonatypeiq_user_token" "test" {
  generated_at = "2026-11-08"
}`
}

const defaultProvider = `
provider "sonatypeiq" {
  username = ""
  password = ""
  url      = ""
}`

// #91: a user without administrator permissions can use the provider for what it is allowed to do.
func TestUserTokenResource_UnprivilegedUser_Offline(t *testing.T) {
	f := newUnprivilegedIq(t, "NexusIQ/1.204.2-01", false)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: userTokenConfig(defaultProvider),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sonatypeiq_user_token.test", "user_code", "abc123"),
					resource.TestCheckResourceAttr("sonatypeiq_user_token.test", "pass_code", "def456"),
				),
			},
			{
				Config:   userTokenConfig(defaultProvider),
				PlanOnly: true,
			},
		},
	})

	if calls := f.usersCalls.Load(); calls != 0 {
		t.Errorf("the provider called %d administrator-only endpoint(s) for an unprivileged user", calls)
	}
}

// A reverse proxy that rewrites the Server header must not stop the provider; it warns and carries on.
func TestUserTokenResource_RewrittenServerHeader_Offline(t *testing.T) {
	newUnprivilegedIq(t, "nginx/1.25", false)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: userTokenConfig(defaultProvider),
				Check:  resource.TestCheckResourceAttr("sonatypeiq_user_token.test", "user_code", "abc123"),
			},
		},
	})
}

func TestProvider_RejectedCredentials_Offline(t *testing.T) {
	newUnprivilegedIq(t, "NexusIQ/1.204.2-01", true)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      userTokenConfig(defaultProvider),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)rejected\s+the\s+supplied\s+credentials`),
			},
		},
	})
}

func TestProvider_ValidatesVersionHint_Offline(t *testing.T) {
	newUnprivilegedIq(t, "NexusIQ/1.204.2-01", false)

	for _, hint := range []string{"1.205.4-01", "205"} {
		t.Run(hint, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config:   userTokenConfig(providerWithHint(hint)),
					PlanOnly: true,
					// The token does not exist yet, so there is always something to create: the point is that
					// the hint is accepted and the plan is produced.
					ExpectNonEmptyPlan: true,
				}},
			})
		})
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      userTokenConfig(providerWithHint("two-oh-five")),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile(`(?s)version_hint.*must\s+be\s+the\s+full\s+version`),
		}},
	})
}

func providerWithHint(hint string) string {
	return fmt.Sprintf(`
provider "sonatypeiq" {
  username     = ""
  password     = ""
  url          = ""
  version_hint = "%s"
}`, hint)
}
