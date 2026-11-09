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
	"regexp"
	"sync"

	"testing"

	"terraform-provider-sonatypeiq/internal/provider/common"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// fakeJiraIqServer is a minimal fake IQ Server for testing the Jira configuration resource.
// It behaves like the real one: PUT merges fields, GET does not return password.
type fakeJiraIqServer struct {
	*httptest.Server
	mu      sync.Mutex
	config  map[string]interface{}
	deleted bool
}

func newFakeJiraIqServer(t *testing.T) *fakeJiraIqServer {
	t.Helper()
	f := &fakeJiraIqServer{
		config: make(map[string]interface{}),
	}

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

func (f *fakeJiraIqServer) route(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/ping":
		return // anonymous
	case "/api/v2/userTokens/currentUser/hasToken":
		http.Error(w, "no token", http.StatusNotFound)
		return
	case "/api/v2/config/jira":
		switch r.Method {
		case http.MethodPut:
			f.handlePut(w, r)
		case http.MethodGet:
			f.handleGet(w)
		case http.MethodDelete:
			f.handleDelete(w)
		default:
			http.NotFound(w, r)
		}
	default:
		http.NotFound(w, r)
	}
}

// handlePut merges fields like the real API does
// handlePut merges fields like the real API does
func (f *fakeJiraIqServer) handlePut(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// After a PUT, the resource exists again (even after DELETE)
	f.deleted = false
	// Merge: only update fields that are present in the request
	for k, v := range body {
		// Normalize customFields to compact JSON (like the real server does with sorted keys)
		if k == "customFields" {
			if v != nil {
				// Re-marshal to get consistent format
				normalized, err := json.Marshal(v)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				var normalizedMap map[string]interface{}
				if err := json.Unmarshal(normalized, &normalizedMap); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				f.config[k] = normalizedMap
			} else {
				f.config[k] = v
			}
		} else {
			f.config[k] = v
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (f *fakeJiraIqServer) handleGet(w http.ResponseWriter) {
	if f.deleted || len(f.config) == 0 {
		http.Error(w, "JIRA not configured.", http.StatusNotFound)
		return
	}
	// Return config without password (as real API does)
	resp := make(map[string]interface{})
	for k, v := range f.config {
		if k != "password" {
			resp[k] = v
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (f *fakeJiraIqServer) handleDelete(w http.ResponseWriter) {
	f.deleted = true
	f.config = make(map[string]interface{})
	w.WriteHeader(http.StatusNoContent)
}

const defaultProvider = `
provider "sonatypeiq" {
  username = ""
  password = ""
  url      = ""
}`

// Test: No credentials - should work
func TestConfigJiraResource_NoCredentials_Offline(t *testing.T) {
	_ = newFakeJiraIqServer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url = "https://jira.example.com"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sonatypeiq_config_jira.test", "id", common.STATE_ID_JIRA_CONFIGURATION),
					resource.TestCheckResourceAttr("sonatypeiq_config_jira.test", "url", "https://jira.example.com"),
					resource.TestCheckNoResourceAttr("sonatypeiq_config_jira.test", "username"),
					resource.TestCheckNoResourceAttr("sonatypeiq_config_jira.test", "password"),
				),
			},
		},
	})
}

// Test: custom_fields that is not a JSON object of objects fails at plan time
func TestConfigJiraResource_InvalidCustomFields_Offline(t *testing.T) {
	_ = newFakeJiraIqServer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url           = "https://jira.example.com"
  custom_fields = "{ not json }"
}`,
				ExpectError: regexp.MustCompile(`(?s)Invalid custom_fields`),
			},
		},
	})
}

// Test: Username alone fails at plan time
func TestConfigJiraResource_UsernameOnly_Offline(t *testing.T) {
	_ = newFakeJiraIqServer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url      = "https://jira.example.com"
  username = "test-user"
}`,
				ExpectError: regexp.MustCompile(`(?s)password.*must.*be.*specified`),
			},
		},
	})
}

// Test: Password alone fails at plan time
func TestConfigJiraResource_PasswordOnly_Offline(t *testing.T) {
	_ = newFakeJiraIqServer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url      = "https://jira.example.com"
  password = "test-password"
}`,
				ExpectError: regexp.MustCompile(`(?s)username.*must.*be.*specified`),
			},
		},
	})
}

// Test: Normal operation with credentials
func TestConfigJiraResource_WithCredentials_Offline(t *testing.T) {
	_ = newFakeJiraIqServer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url      = "https://jira.example.com"
  username = "test-user"
  password = "test-password"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sonatypeiq_config_jira.test", "url", "https://jira.example.com"),
					resource.TestCheckResourceAttr("sonatypeiq_config_jira.test", "username", "test-user"),
					resource.TestCheckResourceAttr("sonatypeiq_config_jira.test", "password", "test-password"),
				),
			},
		},
	})
}

// Test: Removing credentials from config - must DELETE first
func TestConfigJiraResource_RemoveCredentials_Offline(t *testing.T) {
	_ = newFakeJiraIqServer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url      = "https://jira.example.com"
  username = "test-user"
  password = "test-password"
}`,
			},
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url = "https://jira.example.com"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("sonatypeiq_config_jira.test", "username"),
					resource.TestCheckNoResourceAttr("sonatypeiq_config_jira.test", "password"),
				),
			},
		},
	})
}

// Test: Removing custom_fields from config - must DELETE first
func TestConfigJiraResource_RemoveCustomFields_Offline(t *testing.T) {
	_ = newFakeJiraIqServer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url           = "https://jira.example.com"
  username      = "test-user"
  password      = "test-password"
  custom_fields = jsonencode({field1 = {key1 = "value1"}})
}`,
			},
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url      = "https://jira.example.com"
  username = "test-user"
  password = "test-password"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("sonatypeiq_config_jira.test", "custom_fields"),
				),
			},
		},
	})
}

// Test: custom_fields with heredoc and different key order - no diff
func TestConfigJiraResource_CustomFieldsHeredoc_Offline(t *testing.T) {
	_ = newFakeJiraIqServer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url      = "https://jira.example.com"
  username = "test-user"
  password = "test-password"
  custom_fields = <<EOF
{
  "zebra": {"key": "value"},
  "alpha": {"key": "value"}
}
EOF
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sonatypeiq_config_jira.test", "custom_fields"),
				),
			},
			// No-op plan - heredoc must not cause a diff
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url      = "https://jira.example.com"
  username = "test-user"
  password = "test-password"
  custom_fields = <<EOF
{
  "zebra": {"key": "value"},
  "alpha": {"key": "value"}
}
EOF
}`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// Test: Update custom_fields with different key order - CHANGE FORMAT causes diff (expected)
// Then update to content that differs only in server normalization - no diff AFTER apply
func TestConfigJiraResource_CustomFieldsUpdate_Offline(t *testing.T) {
	_ = newFakeJiraIqServer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with jsonencode
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url      = "https://jira.example.com"
  username = "test-user"
  password = "test-password"
  custom_fields = jsonencode({
    zebra = { key = "value" }
    alpha = { key = "value" }
  })
}`,
			},
			// Step 2: Update to heredoc (this WILL show a diff because format changed)
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url      = "https://jira.example.com"
  username = "test-user"
  password = "test-password"
  custom_fields = <<EOF
{
  "alpha": {"key": "value"},
  "zebra": {"key": "value"}
}
EOF
}`,
				Check: resource.TestCheckResourceAttrSet("sonatypeiq_config_jira.test", "custom_fields"),
			},
			// Step 3: No-op plan - same heredoc should not cause diff
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url      = "https://jira.example.com"
  username = "test-user"
  password = "test-password"
  custom_fields = <<EOF
{
  "alpha": {"key": "value"},
  "zebra": {"key": "value"}
}
EOF
}`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// Test: Import
func TestConfigJiraResource_Import_Offline(t *testing.T) {
	_ = newFakeJiraIqServer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: defaultProvider + `
resource "sonatypeiq_config_jira" "test" {
  url      = "https://jira.example.com"
  username = "test-user"
  password = "test-password"
}`,
			},
			{
				ResourceName:            "sonatypeiq_config_jira.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "last_updated"},
			},
		},
	})
}
