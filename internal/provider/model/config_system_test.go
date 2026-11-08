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
package model_test

import (
	"terraform-provider-sonatypeiq/internal/provider/model"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// emptyResource returns a resource model with every optional property unset (null), as Terraform would.
func emptyResource() model.SystemConfigResource {
	var m model.SystemConfigResource
	m.BaseURL = types.StringValue("https://iq.example.com/")
	m.ForceBaseURL = types.BoolValue(false)
	m.AdvancedReportingInsightsEnabled = types.BoolNull()
	m.AdvancedSearchCsvExportDelimiter = types.StringNull()
	m.AlpObservedLicenseDetectionEnabled = types.BoolNull()
	m.ApiAccessAllowList = types.SetNull(types.StringType)
	m.CsrfProtection = types.BoolNull()
	m.EventBusMaxThreadPoolSize = types.Int32Null()
	m.FrameAncestorsAllowlist = types.SetNull(types.StringType)
	m.HdsURL = types.StringNull()
	m.IntegrationsSupportedVersionCount = types.Int32Null()
	m.MaxAdvancedSearchClauseCount = types.Int32Null()
	m.PolicyMonitoringHour = types.Int32Null()
	m.PurgeScanFiles = types.StringNull()
	m.SessionTimeout = types.Int32Null()
	m.SuccessMetricsStageID = types.StringNull()
	m.UserAgentSuffix = types.StringNull()
	m.WaivedComponentUpgradeMonitoringEnabled = types.BoolNull()
	m.WebhookSecretPassphrase = types.StringNull()
	return m
}

// Every property the API client knows about must be registered, so that it is gated and documented.
func TestSystemConfigPropertiesCoverTheClient(t *testing.T) {
	registered := map[sonatypeiq.SystemConfigProperty]bool{}
	for _, p := range model.SystemConfigProperties {
		registered[p.ApiProperty] = true
	}
	for _, p := range sonatypeiq.AllowedSystemConfigPropertyEnumValues {
		if !registered[p] {
			t.Errorf("API property %q is not registered in model.SystemConfigProperties - add it with the release that introduced it", p)
		}
	}
	if len(registered) != len(model.SystemConfigProperties) {
		t.Errorf("model.SystemConfigProperties contains duplicate API properties")
	}
}

func TestSystemConfigIntroducedInReleases(t *testing.T) {
	// Established from the Sonatype IQ Server release history; Sonatype's published documentation
	// (https://help.sonatype.com/en/configuration-rest-api.html) agrees wherever it states a release, except for
	// alp_observed_license_detection_enabled, which it gives as 165.
	expected := map[string]int32{
		"base_url":                                    138,
		"force_base_url":                              138,
		"advanced_search_csv_export_delimiter":        142,
		"csrf_protection":                             142,
		"event_bus_max_thread_pool_size":              142,
		"frame_ancestors_allowlist":                   142,
		"hds_url":                                     142,
		"max_advanced_search_clause_count":            142,
		"policy_monitoring_hour":                      142,
		"user_agent_suffix":                           142,
		"webhook_secret_passphrase":                   142,
		"session_timeout":                             143,
		"purge_scan_files":                            152,
		"alp_observed_license_detection_enabled":      161,
		"waived_component_upgrade_monitoring_enabled": 162,
		"advanced_reporting_insights_enabled":         175,
		"api_access_allow_list":                       175,
		"success_metrics_stage_id":                    182,
		"integrations_supported_version_count":        197,
	}
	if len(expected) != len(model.SystemConfigProperties) {
		t.Errorf("expected %d properties, registry has %d", len(expected), len(model.SystemConfigProperties))
	}
	for attribute, release := range expected {
		p, ok := model.SystemConfigPropertyByAttribute(attribute)
		if !ok {
			t.Errorf("%s is not registered", attribute)
			continue
		}
		if p.IntroducedIn != release {
			t.Errorf("%s: introduced in %d, want %d", attribute, p.IntroducedIn, release)
		}
	}
}

func TestSystemConfigUnsupportedProperties(t *testing.T) {
	m := emptyResource()
	if got := m.UnsupportedProperties(190); len(got) != 0 {
		t.Fatalf("no optional properties set, but got unsupported: %v", got)
	}

	m.IntegrationsSupportedVersionCount = types.Int32Value(5) // 197
	m.SuccessMetricsStageID = types.StringValue("build")      // 182
	m.SessionTimeout = types.Int32Value(60)                   // 143

	got := m.UnsupportedProperties(190)
	if len(got) != 1 || got[0].Attribute != "integrations_supported_version_count" {
		t.Fatalf("on 190 only integrations_supported_version_count should be unsupported, got %v", got)
	}
	if len(m.UnsupportedProperties(197)) != 0 {
		t.Errorf("nothing should be unsupported on 197")
	}
	if len(m.UnsupportedProperties(0)) != 0 {
		t.Errorf("an undetected server version must never block a property")
	}
	if len(m.UnsupportedProperties(181)) != 2 {
		t.Errorf("on 181 both 182 and 197 properties should be unsupported, got %v", m.UnsupportedProperties(181))
	}
	if len(m.UnsupportedProperties(142)) != 3 {
		t.Errorf("on 142 the 143, 182 and 197 properties should be unsupported, got %v", m.UnsupportedProperties(142))
	}
}

func TestSystemConfigUnsupportedMessageNamesTheVersions(t *testing.T) {
	p, _ := model.SystemConfigPropertyByAttribute("integrations_supported_version_count")
	msg := p.UnsupportedMessage(190)
	for _, want := range []string{"integrations_supported_version_count", "197", "190"} {
		if !contains(msg, want) {
			t.Errorf("message %q does not mention %q", msg, want)
		}
	}
}

func TestSystemConfigMapToApiOnlySendsSetProperties(t *testing.T) {
	m := emptyResource()
	m.SessionTimeout = types.Int32Value(45)
	m.ApiAccessAllowList = stringSet("alice", "bob")

	api := m.MapToApi()
	if api.SessionTimeout.Get() == nil || *api.SessionTimeout.Get() != 45 {
		t.Errorf("session timeout not mapped")
	}
	if len(api.ApiAccessAllowList) != 2 || api.ApiAccessAllowList[0] != "alice" {
		t.Errorf("api access allow list not mapped: %v", api.ApiAccessAllowList)
	}
	if api.CsrfProtection.IsSet() || api.HdsUrl.IsSet() || api.PurgeScanFiles.IsSet() {
		t.Errorf("properties that are not set must not be sent")
	}
}

func TestSystemConfigRemovedProperties(t *testing.T) {
	state := emptyResource()
	state.SessionTimeout = types.Int32Value(45)
	state.UserAgentSuffix = types.StringValue("Approved")

	plan := emptyResource()
	plan.SessionTimeout = types.Int32Value(60)

	removed := state.RemovedProperties(&plan)
	if len(removed) != 1 || removed[0].Attribute != "user_agent_suffix" {
		t.Fatalf("only user_agent_suffix was removed, got %v", removed)
	}
}

func TestSystemConfigMapFromApiLeavesUnmanagedPropertiesNull(t *testing.T) {
	api := sonatypeiq.NewSystemConfigWithDefaults()
	api.BaseUrl = *sonatypeiq.NewNullableString(strPtr("https://iq.example.com/"))
	api.ForceBaseUrl = *sonatypeiq.NewNullableBool(boolPtr(false))
	// Sonatype IQ Server reports defaults for properties that Terraform does not manage
	api.SessionTimeout = *sonatypeiq.NewNullableInt32(int32Ptr(30))
	api.CsrfProtection = *sonatypeiq.NewNullableBool(boolPtr(true))
	api.PolicyMonitoringHour = *sonatypeiq.NewNullableInt32(int32Ptr(3))

	m := emptyResource()
	m.PolicyMonitoringHour = types.Int32Value(3) // managed
	m.WebhookSecretPassphrase = types.StringValue("s3cret")

	m.MapFromApi(api)

	if !m.SessionTimeout.IsNull() || !m.CsrfProtection.IsNull() {
		t.Errorf("unmanaged properties must stay null so server defaults are not drift")
	}
	if m.PolicyMonitoringHour.ValueInt32() != 3 {
		t.Errorf("managed property not refreshed from the API")
	}
	if m.WebhookSecretPassphrase.ValueString() != "s3cret" {
		t.Errorf("the webhook secret must be kept from configuration, never read back")
	}
}

func TestRequestedPropertiesForVersion(t *testing.T) {
	has := func(props []sonatypeiq.SystemConfigProperty, want sonatypeiq.SystemConfigProperty) bool {
		for _, p := range props {
			if p == want {
				return true
			}
		}
		return false
	}
	old := model.RequestedPropertiesFor(190)
	if has(old, sonatypeiq.INTEGRATIONS_SUPPORTED_VERSION_COUNT) {
		t.Errorf("a 190 server must not be asked for a property introduced in 197")
	}
	if !has(old, sonatypeiq.BASE_URL) || !has(old, sonatypeiq.SUCCESS_METRICS_STAGE_ID) {
		t.Errorf("a 190 server supports baseUrl and successMetricsStageId")
	}
	if !has(model.RequestedPropertiesFor(205), sonatypeiq.INTEGRATIONS_SUPPORTED_VERSION_COUNT) {
		t.Errorf("a 205 server supports integrationsSupportedVersionCount")
	}
}

func stringSet(values ...string) types.Set {
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	return types.SetValueMust(types.StringType, elements)
}

func apiFromServer() *sonatypeiq.SystemConfig {
	api := sonatypeiq.NewSystemConfigWithDefaults()
	api.BaseUrl = *sonatypeiq.NewNullableString(strPtr("https://iq.example.com/"))
	api.ForceBaseUrl = *sonatypeiq.NewNullableBool(boolPtr(false))
	return api
}

func setElements(set types.Set) []string {
	var out []string
	for _, e := range set.Elements() {
		out = append(out, e.(types.String).ValueString())
	}
	return out
}

// Sonatype IQ Server always adds 'self' to a frame ancestors allowlist; it must not look like drift.
func TestSystemConfigFrameAncestorsIgnoresTheImpliedSelf(t *testing.T) {
	api := apiFromServer()
	api.FrameAncestorsAllowlist = []string{"'self'", "*.example.com"}

	m := emptyResource()
	m.FrameAncestorsAllowlist = stringSet("*.example.com")
	m.MapFromApi(api)

	if got := setElements(m.FrameAncestorsAllowlist); len(got) != 1 || got[0] != "*.example.com" {
		t.Errorf("'self' must be hidden when it is not configured, got %v", got)
	}
}

func TestSystemConfigFrameAncestorsKeepsSelfWhenConfigured(t *testing.T) {
	api := apiFromServer()
	api.FrameAncestorsAllowlist = []string{"'self'", "*.example.com"}

	m := emptyResource()
	m.FrameAncestorsAllowlist = stringSet("*.example.com", "'self'") // order is irrelevant for a set
	m.MapFromApi(api)

	if got := setElements(m.FrameAncestorsAllowlist); len(got) != 2 {
		t.Errorf("a configured 'self' must be kept, got %v", got)
	}
	if !m.FrameAncestorsAllowlist.Equal(stringSet("'self'", "*.example.com")) {
		t.Errorf("the set must equal the configured values regardless of order")
	}
}

func TestSystemConfigFrameAncestorsNoneStaysAlone(t *testing.T) {
	api := apiFromServer()
	api.FrameAncestorsAllowlist = []string{"'none'"}

	m := emptyResource()
	m.FrameAncestorsAllowlist = stringSet("'none'")
	m.MapFromApi(api)

	if !m.FrameAncestorsAllowlist.Equal(stringSet("'none'")) {
		t.Errorf("'none' must round-trip unchanged, got %v", setElements(m.FrameAncestorsAllowlist))
	}
}

func TestSystemConfigFrameAncestorsUnmanagedStaysNull(t *testing.T) {
	api := apiFromServer()
	api.FrameAncestorsAllowlist = []string{"'self'"}

	m := emptyResource()
	m.MapFromApi(api)

	if !m.FrameAncestorsAllowlist.IsNull() {
		t.Errorf("an unmanaged allowlist must stay null")
	}
}

// Sonatype IQ Server de-duplicates the API access list and does not keep its order.
func TestSystemConfigApiAccessAllowListIgnoresOrder(t *testing.T) {
	api := apiFromServer()
	api.ApiAccessAllowList = []string{"bob", "alice"}

	m := emptyResource()
	m.ApiAccessAllowList = stringSet("alice", "bob")
	m.MapFromApi(api)

	if !m.ApiAccessAllowList.Equal(stringSet("alice", "bob")) {
		t.Errorf("order must not matter, got %v", setElements(m.ApiAccessAllowList))
	}
}

// Sonatype IQ Server always adds a trailing slash to URLs.
func TestSystemConfigUrlsIgnoreTheTrailingSlashTheServerAdds(t *testing.T) {
	api := apiFromServer()
	api.BaseUrl = *sonatypeiq.NewNullableString(strPtr("https://iq.example.com/"))
	api.HdsUrl = *sonatypeiq.NewNullableString(strPtr("https://hds.example.com/"))

	m := emptyResource()
	m.BaseURL = types.StringValue("https://iq.example.com")
	m.HdsURL = types.StringValue("https://hds.example.com")
	m.MapFromApi(api)

	if m.BaseURL.ValueString() != "https://iq.example.com" || m.HdsURL.ValueString() != "https://hds.example.com" {
		t.Errorf("a configured URL without the slash must be kept, got %q and %q", m.BaseURL.ValueString(), m.HdsURL.ValueString())
	}
}

func TestSystemConfigBaseUrlIsReadOnImportOrWhenItReallyChanged(t *testing.T) {
	api := apiFromServer()

	imported := model.SystemConfigResource{}
	imported.BaseURL = types.StringNull()
	imported.MapFromApi(api)
	if imported.BaseURL.ValueString() != "https://iq.example.com/" {
		t.Errorf("on import the base URL must come from the server, got %q", imported.BaseURL.ValueString())
	}

	m := emptyResource()
	m.BaseURL = types.StringValue("https://old.example.com")
	m.MapFromApi(api)
	if m.BaseURL.ValueString() != "https://iq.example.com/" {
		t.Errorf("a genuinely different base URL must be reported, got %q", m.BaseURL.ValueString())
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func int32Ptr(i int32) *int32 { return &i }
