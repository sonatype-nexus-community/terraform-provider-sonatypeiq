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
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// fakeIqServer is just enough of Sonatype IQ Server to drive sonatypeiq_system_config end to end through the
// Terraform CLI without a licensed server: it reports a chosen release in its Server header, stores the
// properties it is sent, reports a server-side default for properties that were never set (as the real
// server does), and records which properties were requested and deleted.
type fakeIqServer struct {
	*httptest.Server
	mu        sync.Mutex
	store     map[string]any
	requested map[string]bool
	deleted   []string
}

var fakeIqDefaults = map[string]any{
	"sessionTimeout":       float64(30),
	"csrfProtection":       true,
	"hdsUrl":               "https://clm.sonatype.com/",
	"policyMonitoringHour": float64(0),
}

func newFakeIqServer(t *testing.T, release int) *fakeIqServer {
	t.Helper()
	return newFakeIqServerWithHeader(t, fmt.Sprintf("NexusIQ/1.%d.0-01", release))
}

// newFakeIqServerWithHeader reports the given Server header, e.g. one rewritten by a reverse proxy.
func newFakeIqServerWithHeader(t *testing.T, serverHeader string) *fakeIqServer {
	t.Helper()
	f := &fakeIqServer{store: map[string]any{}, requested: map[string]bool{}}

	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", serverHeader)
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
func (f *fakeIqServer) route(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/ping":
		return // anonymous
	case r.URL.Path == "/api/v2/userTokens/currentUser/hasToken":
		http.Error(w, "no token", http.StatusNotFound) // the credentials check: reachable and accepted
		return
	case strings.HasPrefix(r.URL.Path, "/api/v2/users/"):
		f.handleUser(w)
		return
	}
	if r.URL.Path != "/api/v2/config" {
		http.NotFound(w, r)
		return
	}

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
}

func (f *fakeIqServer) handleUser(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"username":"admin"}`))
}

func (f *fakeIqServer) handlePut(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for k, v := range body {
		f.storeNormalised(k, v)
	}
	w.WriteHeader(http.StatusNoContent)
}

// storeNormalised stores a property the way Sonatype IQ Server does, including the normalisation it applies.
func (f *fakeIqServer) storeNormalised(property string, value any) {
	switch property {
	case "baseUrl", "hdsUrl":
		// A trailing slash is always added
		if s, ok := value.(string); ok && !strings.HasSuffix(s, "/") {
			value = s + "/"
		}
	case "frameAncestorsAllowlist":
		value = normaliseFrameAncestors(toStrings(value))
	case "apiAccessAllowList":
		// Duplicates are removed (through a hash set) and the order is not kept
		value = reverse(distinct(toStrings(value)))
	}
	if list, ok := value.([]string); ok && len(list) == 0 {
		delete(f.store, property)
		return
	}
	f.store[property] = value
}

func toStrings(value any) []string {
	var out []string
	items, _ := value.([]any)
	for _, i := range items {
		if s, ok := i.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func distinct(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func reverse(values []string) []string {
	out := make([]string, 0, len(values))
	for i := len(values) - 1; i >= 0; i-- {
		out = append(out, values[i])
	}
	return out
}

// normaliseFrameAncestors mirrors Sonatype IQ Server: 'none' is stored alone, otherwise 'self' is always added first.
func normaliseFrameAncestors(values []string) []string {
	for _, v := range values {
		if v == "'none'" {
			return []string{"'none'"}
		}
	}
	if len(values) == 0 {
		return nil
	}
	return distinct(append([]string{"'self'"}, values...))
}

func (f *fakeIqServer) handleGet(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{}
	for _, p := range propertyNames(r) {
		f.requested[p] = true
		if v, ok := f.store[p]; ok {
			resp[p] = v
		} else if v, ok := fakeIqDefaults[p]; ok {
			resp[p] = v
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (f *fakeIqServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	for _, p := range propertyNames(r) {
		f.deleted = append(f.deleted, p)
		delete(f.store, p)
	}
	w.WriteHeader(http.StatusNoContent)
}

func propertyNames(r *http.Request) []string {
	var names []string
	for _, v := range r.URL.Query()["property"] {
		names = append(names, strings.Split(v, ",")...)
	}
	return names
}

func (f *fakeIqServer) checkStored(property string, want any) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		f.mu.Lock()
		defer f.mu.Unlock()
		if got, ok := f.store[property]; !ok || fmt.Sprint(got) != fmt.Sprint(want) {
			return fmt.Errorf("server has %s = %v (set: %t), want %v", property, got, ok, want)
		}
		return nil
	}
}

func (f *fakeIqServer) checkNeverRequested(property string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.requested[property] {
			return fmt.Errorf("%s was requested from a server that is too old to know it", property)
		}
		if _, ok := f.store[property]; ok {
			return fmt.Errorf("%s was sent to a server that is too old to know it", property)
		}
		return nil
	}
}

func (f *fakeIqServer) checkDeleted(property string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		f.mu.Lock()
		defer f.mu.Unlock()
		for _, d := range f.deleted {
			if d == property {
				return nil
			}
		}
		return fmt.Errorf("%s was not deleted, deleted: %v", property, f.deleted)
	}
}

func fakeSystemConfig(extra string) string {
	return utils_test.ProviderConfig + `
resource "sonatypeiq_system_config" "test" {
  base_url       = "https://iq.example.com"
  force_base_url = false
` + extra + `
}`
}

// A server older than the release that introduced a property must refuse it at plan time, name the
// releases involved, and never be sent or asked for the property.
func TestSystemConfigResource_GatesPropertiesByRelease_Offline(t *testing.T) {
	f := newFakeIqServer(t, 190)
	resourceName := "sonatypeiq_system_config.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Supported on 190: success_metrics_stage_id (182) and an undocumented-release property
			{
				Config: fakeSystemConfig(`
  session_timeout          = 45
  success_metrics_stage_id = "build"
  webhook_secret_passphrase = "s3cret"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "session_timeout", "45"),
					resource.TestCheckResourceAttr(resourceName, "success_metrics_stage_id", "build"),
					// The server reports a default of 30 / true for these, which must not appear as drift
					resource.TestCheckNoResourceAttr(resourceName, "csrf_protection"),
					resource.TestCheckNoResourceAttr(resourceName, "policy_monitoring_hour"),
					resource.TestCheckResourceAttr(resourceName, "webhook_secret_passphrase", "s3cret"),
					f.checkStored("sessionTimeout", 45),
					f.checkStored("successMetricsStageId", "build"),
					f.checkNeverRequested("integrationsSupportedVersionCount"),
				),
			},
			// No drift
			{
				Config: fakeSystemConfig(`
  session_timeout          = 45
  success_metrics_stage_id = "build"
  webhook_secret_passphrase = "s3cret"`),
				PlanOnly: true,
			},
			// Introduced in 197: refused on 190, with an explanation
			{
				Config: fakeSystemConfig(`
  session_timeout                      = 45
  integrations_supported_version_count = 5`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)introduced in Sonatype IQ Server\s+release\s+197.*connected\s+server\s+is\s+release\s+190`),
			},
			// Removing a property deletes it so that the server reverts to its default
			{
				Config: fakeSystemConfig(`
  success_metrics_stage_id = "build"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "session_timeout"),
					f.checkDeleted("sessionTimeout"),
					f.checkNeverRequested("integrationsSupportedVersionCount"),
				),
			},
		},
	})
}

func TestSystemConfigResource_AcceptsNewPropertiesOnNewReleases_Offline(t *testing.T) {
	f := newFakeIqServer(t, 205)
	resourceName := "sonatypeiq_system_config.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fakeSystemConfig(`
  integrations_supported_version_count = 5
  api_access_allow_list                = ["alice", "bob"]
  frame_ancestors_allowlist            = ["*.example.com"]
  hds_url                              = "https://hds.example.com"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "integrations_supported_version_count", "5"),
					resource.TestCheckResourceAttr(resourceName, "api_access_allow_list.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "api_access_allow_list.*", "alice"),
					resource.TestCheckTypeSetElemAttr(resourceName, "api_access_allow_list.*", "bob"),
					// 'self' is added by the server but is not part of the configuration
					resource.TestCheckResourceAttr(resourceName, "frame_ancestors_allowlist.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "frame_ancestors_allowlist.*", "*.example.com"),
					// The server adds a trailing slash; the configured value is kept
					resource.TestCheckResourceAttr(resourceName, "base_url", "https://iq.example.com"),
					resource.TestCheckResourceAttr(resourceName, "hds_url", "https://hds.example.com"),
					f.checkStored("integrationsSupportedVersionCount", 5),
					f.checkStored("frameAncestorsAllowlist", []string{"'self'", "*.example.com"}),
					f.checkStored("baseUrl", "https://iq.example.com/"),
				),
			},
			{
				Config: fakeSystemConfig(`
  integrations_supported_version_count = 5
  api_access_allow_list                = ["alice", "bob"]
  frame_ancestors_allowlist            = ["*.example.com"]
  hds_url                              = "https://hds.example.com"`),
				PlanOnly: true,
			},
			// 'self' may be configured explicitly, and the order is irrelevant
			{
				Config: fakeSystemConfig(`
  api_access_allow_list     = ["bob", "alice"]
  frame_ancestors_allowlist = ["*.example.com", "'self'"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "frame_ancestors_allowlist.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "frame_ancestors_allowlist.*", "'self'"),
				),
			},
			{
				Config: fakeSystemConfig(`
  api_access_allow_list     = ["bob", "alice"]
  frame_ancestors_allowlist = ["*.example.com", "'self'"]`),
				PlanOnly: true,
			},
			// 'none' forbids all framing and stands alone
			{
				Config: fakeSystemConfig(`frame_ancestors_allowlist = ["'none'"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "frame_ancestors_allowlist.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "frame_ancestors_allowlist.*", "'none'"),
				),
			},
		},
	})
}

// Values outside the documented range are rejected before anything reaches the server.
func TestSystemConfigResource_ValidatesValues_Offline(t *testing.T) {
	newFakeIqServer(t, 205)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      fakeSystemConfig(`policy_monitoring_hour = 24`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)policy_monitoring_hour.*(between|0 and 23)`),
			},
			{
				Config:      fakeSystemConfig(`frame_ancestors_allowlist = ["'none'", "*.example.com"]`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)Invalid frame ancestors allowlist`),
			},
			{
				Config:      fakeSystemConfig(`session_timeout = 2`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)session_timeout.*(between|3 and 120)`),
			},
			{
				Config:      fakeSystemConfig(`api_access_allow_list = []`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)api_access_allow_list.*(at least|1)`),
			},
			{
				Config:      fakeSystemConfig(`success_metrics_stage_id = "nonsense"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)success_metrics_stage_id.*(one of|must be)`),
			},
		},
	})
}

func fakeSystemConfigWithHint(hint, extra string) string {
	providerBlock := `
provider "sonatypeiq" {
  username = ""
  password = ""
  url      = ""
`
	if hint != "" {
		providerBlock += fmt.Sprintf("  version_hint = %q\n", hint)
	}
	providerBlock += "}\n"

	return providerBlock + `
resource "sonatypeiq_system_config" "test" {
  base_url       = "https://iq.example.com"
  force_base_url = false
` + extra + `
}`
}

// Behind a reverse proxy that rewrites the Server header the version is unknown: nothing is blocked, and the
// provider still works.
func TestSystemConfigResource_RewrittenServerHeaderIsNotBlocking_Offline(t *testing.T) {
	f := newFakeIqServerWithHeader(t, "nginx/1.25")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fakeSystemConfigWithHint("", `integrations_supported_version_count = 5`),
				Check:  f.checkStored("integrationsSupportedVersionCount", 5),
			},
		},
	})
}

// version_hint switches the release checks back on when the Server header is rewritten.
func TestSystemConfigResource_VersionHintRestoresGating_Offline(t *testing.T) {
	newFakeIqServerWithHeader(t, "nginx/1.25")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      fakeSystemConfigWithHint("1.190.0-01", `integrations_supported_version_count = 5`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)introduced in Sonatype IQ Server\s+release\s+197.*connected\s+server\s+is\s+release\s+190`),
			},
		},
	})
}

// The hint takes precedence over what the server reports.
func TestSystemConfigResource_VersionHintOverridesTheServer_Offline(t *testing.T) {
	f := newFakeIqServer(t, 190)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fakeSystemConfigWithHint("205", `integrations_supported_version_count = 5`),
				Check:  f.checkStored("integrationsSupportedVersionCount", 5),
			},
		},
	})
}
