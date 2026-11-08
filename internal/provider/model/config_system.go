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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// SystemConfigModel
// ------------------------------------------------------------
type SystemConfigModel struct {
	ID                                      types.String `tfsdk:"id"`
	BaseURL                                 types.String `tfsdk:"base_url"`
	ForceBaseURL                            types.Bool   `tfsdk:"force_base_url"`
	AdvancedReportingInsightsEnabled        types.Bool   `tfsdk:"advanced_reporting_insights_enabled"`
	AdvancedSearchCsvExportDelimiter        types.String `tfsdk:"advanced_search_csv_export_delimiter"`
	AlpObservedLicenseDetectionEnabled      types.Bool   `tfsdk:"alp_observed_license_detection_enabled"`
	ApiAccessAllowList                      types.List   `tfsdk:"api_access_allow_list"`
	CsrfProtection                          types.Bool   `tfsdk:"csrf_protection"`
	EventBusMaxThreadPoolSize               types.Int32  `tfsdk:"event_bus_max_thread_pool_size"`
	FrameAncestorsAllowlist                 types.List   `tfsdk:"frame_ancestors_allowlist"`
	HdsURL                                  types.String `tfsdk:"hds_url"`
	IntegrationsSupportedVersionCount       types.Int32  `tfsdk:"integrations_supported_version_count"`
	MaxAdvancedSearchClauseCount            types.Int32  `tfsdk:"max_advanced_search_clause_count"`
	PolicyMonitoringHour                    types.Int32  `tfsdk:"policy_monitoring_hour"`
	PurgeScanFiles                          types.String `tfsdk:"purge_scan_files"`
	SessionTimeout                          types.Int32  `tfsdk:"session_timeout"`
	SuccessMetricsStageID                   types.String `tfsdk:"success_metrics_stage_id"`
	UserAgentSuffix                         types.String `tfsdk:"user_agent_suffix"`
	WaivedComponentUpgradeMonitoringEnabled types.Bool   `tfsdk:"waived_component_upgrade_monitoring_enabled"`
}

// MapFromApi maps every property returned by the API.
func (m *SystemConfigModel) MapFromApi(api *sonatypeiq.SystemConfig) {
	m.BaseURL = types.StringPointerValue(api.BaseUrl.Get())
	m.ForceBaseURL = types.BoolPointerValue(api.ForceBaseUrl.Get())
	m.AdvancedReportingInsightsEnabled = types.BoolPointerValue(api.ADVANCED_REPORTING_INSIGHTS_ENABLED.Get())
	m.AdvancedSearchCsvExportDelimiter = types.StringPointerValue(api.AdvancedSearchCSVExportDelimiter.Get())
	m.AlpObservedLicenseDetectionEnabled = types.BoolPointerValue(api.AlpObservedLicenseDetectionEnabled.Get())
	m.ApiAccessAllowList = stringListValue(api.ApiAccessAllowList)
	m.CsrfProtection = types.BoolPointerValue(api.CsrfProtection.Get())
	m.EventBusMaxThreadPoolSize = types.Int32PointerValue(api.EventBusMaxThreadPoolSize.Get())
	m.FrameAncestorsAllowlist = stringListValue(api.FrameAncestorsAllowlist)
	m.HdsURL = types.StringPointerValue(api.HdsUrl.Get())
	m.IntegrationsSupportedVersionCount = types.Int32PointerValue(api.IntegrationsSupportedVersionCount.Get())
	m.MaxAdvancedSearchClauseCount = types.Int32PointerValue(api.MaxAdvancedSearchClauseCount.Get())
	m.PolicyMonitoringHour = types.Int32PointerValue(api.PolicyMonitoringHour.Get())
	m.PurgeScanFiles = types.StringPointerValue(api.PurgeScanFiles.Get())
	m.SessionTimeout = types.Int32PointerValue(api.SessionTimeout.Get())
	m.SuccessMetricsStageID = types.StringPointerValue(api.SuccessMetricsStageId.Get())
	m.UserAgentSuffix = types.StringPointerValue(api.UserAgentSuffix.Get())
	m.WaivedComponentUpgradeMonitoringEnabled = types.BoolPointerValue(api.WaivedComponentUpgradeMonitoringEnabled.Get())
}

// RequestedPropertiesFor returns the API properties to read for the given IQ Server release: the base
// properties plus every other property that the release is documented to support.
func RequestedPropertiesFor(iqVersion int32) []sonatypeiq.SystemConfigProperty {
	props := make([]sonatypeiq.SystemConfigProperty, 0, len(SystemConfigProperties))
	for _, p := range SystemConfigProperties {
		if p.IsSupportedBy(iqVersion) {
			props = append(props, p.ApiProperty)
		}
	}
	return props
}

// BaseProperties are the properties that every supported Sonatype IQ Server has always had.
func BaseProperties() []sonatypeiq.SystemConfigProperty {
	return []sonatypeiq.SystemConfigProperty{sonatypeiq.BASE_URL, sonatypeiq.FORCE_BASE_URL}
}

// SystemConfigResource
// ------------------------------------------------------------
type SystemConfigResource struct {
	SystemConfigModel
	WebhookSecretPassphrase types.String `tfsdk:"webhook_secret_passphrase"`
	LastUpdated             types.String `tfsdk:"last_updated"`
}

// attributeValue returns the Terraform value of a managed property by its attribute name.
func (m *SystemConfigResource) attributeValue(attribute string) attr.Value {
	switch attribute {
	case "base_url":
		return m.BaseURL
	case "force_base_url":
		return m.ForceBaseURL
	case "advanced_reporting_insights_enabled":
		return m.AdvancedReportingInsightsEnabled
	case "advanced_search_csv_export_delimiter":
		return m.AdvancedSearchCsvExportDelimiter
	case "alp_observed_license_detection_enabled":
		return m.AlpObservedLicenseDetectionEnabled
	case "api_access_allow_list":
		return m.ApiAccessAllowList
	case "csrf_protection":
		return m.CsrfProtection
	case "event_bus_max_thread_pool_size":
		return m.EventBusMaxThreadPoolSize
	case "frame_ancestors_allowlist":
		return m.FrameAncestorsAllowlist
	case "hds_url":
		return m.HdsURL
	case "integrations_supported_version_count":
		return m.IntegrationsSupportedVersionCount
	case "max_advanced_search_clause_count":
		return m.MaxAdvancedSearchClauseCount
	case "policy_monitoring_hour":
		return m.PolicyMonitoringHour
	case "purge_scan_files":
		return m.PurgeScanFiles
	case "session_timeout":
		return m.SessionTimeout
	case "success_metrics_stage_id":
		return m.SuccessMetricsStageID
	case "user_agent_suffix":
		return m.UserAgentSuffix
	case "waived_component_upgrade_monitoring_enabled":
		return m.WaivedComponentUpgradeMonitoringEnabled
	case "webhook_secret_passphrase":
		return m.WebhookSecretPassphrase
	}
	return nil
}

// IsSet reports whether a property is set (or will be set) in Terraform.
func (m *SystemConfigResource) IsSet(attribute string) bool {
	v := m.attributeValue(attribute)
	return v != nil && !v.IsNull()
}

// PropertiesInUse returns the properties that are set in Terraform, always including the base properties.
func (m *SystemConfigResource) PropertiesInUse() []SystemConfigPropertyInfo {
	var inUse []SystemConfigPropertyInfo
	for _, p := range SystemConfigProperties {
		if m.IsSet(p.Attribute) {
			inUse = append(inUse, p)
		}
	}
	return inUse
}

// ApiPropertiesInUse is PropertiesInUse as the names used by the Configuration REST API.
func (m *SystemConfigResource) ApiPropertiesInUse() []sonatypeiq.SystemConfigProperty {
	inUse := m.PropertiesInUse()
	props := make([]sonatypeiq.SystemConfigProperty, 0, len(inUse))
	for _, p := range inUse {
		props = append(props, p.ApiProperty)
	}
	return props
}

// UnsupportedProperties returns the properties that are set but are not supported by the given
// Sonatype IQ Server release.
func (m *SystemConfigResource) UnsupportedProperties(iqVersion int32) []SystemConfigPropertyInfo {
	var unsupported []SystemConfigPropertyInfo
	for _, p := range m.PropertiesInUse() {
		if !p.IsSupportedBy(iqVersion) {
			unsupported = append(unsupported, p)
		}
	}
	return unsupported
}

// RemovedProperties returns the properties that are set in this (prior state) model but no longer set
// in the plan. They are deleted so that Sonatype IQ Server reverts to its default.
func (m *SystemConfigResource) RemovedProperties(plan *SystemConfigResource) []SystemConfigPropertyInfo {
	var removed []SystemConfigPropertyInfo
	for _, p := range m.PropertiesInUse() {
		if !plan.IsSet(p.Attribute) {
			removed = append(removed, p)
		}
	}
	return removed
}

// MapToApi maps only the properties that are set, so unset properties keep their server default.
func (m *SystemConfigResource) MapToApi() *sonatypeiq.SystemConfig {
	api := sonatypeiq.NewSystemConfigWithDefaults()
	if m.IsSet("base_url") {
		api.BaseUrl = *sonatypeiq.NewNullableString(m.BaseURL.ValueStringPointer())
	}
	if m.IsSet("force_base_url") {
		api.ForceBaseUrl = *sonatypeiq.NewNullableBool(m.ForceBaseURL.ValueBoolPointer())
	}
	if m.IsSet("advanced_reporting_insights_enabled") {
		api.ADVANCED_REPORTING_INSIGHTS_ENABLED = *sonatypeiq.NewNullableBool(m.AdvancedReportingInsightsEnabled.ValueBoolPointer())
	}
	if m.IsSet("advanced_search_csv_export_delimiter") {
		api.AdvancedSearchCSVExportDelimiter = *sonatypeiq.NewNullableString(m.AdvancedSearchCsvExportDelimiter.ValueStringPointer())
	}
	if m.IsSet("alp_observed_license_detection_enabled") {
		api.AlpObservedLicenseDetectionEnabled = *sonatypeiq.NewNullableBool(m.AlpObservedLicenseDetectionEnabled.ValueBoolPointer())
	}
	if m.IsSet("api_access_allow_list") {
		api.ApiAccessAllowList = stringsFromList(m.ApiAccessAllowList)
	}
	if m.IsSet("csrf_protection") {
		api.CsrfProtection = *sonatypeiq.NewNullableBool(m.CsrfProtection.ValueBoolPointer())
	}
	if m.IsSet("event_bus_max_thread_pool_size") {
		api.EventBusMaxThreadPoolSize = *sonatypeiq.NewNullableInt32(m.EventBusMaxThreadPoolSize.ValueInt32Pointer())
	}
	if m.IsSet("frame_ancestors_allowlist") {
		api.FrameAncestorsAllowlist = stringsFromList(m.FrameAncestorsAllowlist)
	}
	if m.IsSet("hds_url") {
		api.HdsUrl = *sonatypeiq.NewNullableString(m.HdsURL.ValueStringPointer())
	}
	if m.IsSet("integrations_supported_version_count") {
		api.IntegrationsSupportedVersionCount = *sonatypeiq.NewNullableInt32(m.IntegrationsSupportedVersionCount.ValueInt32Pointer())
	}
	if m.IsSet("max_advanced_search_clause_count") {
		api.MaxAdvancedSearchClauseCount = *sonatypeiq.NewNullableInt32(m.MaxAdvancedSearchClauseCount.ValueInt32Pointer())
	}
	if m.IsSet("policy_monitoring_hour") {
		api.PolicyMonitoringHour = *sonatypeiq.NewNullableInt32(m.PolicyMonitoringHour.ValueInt32Pointer())
	}
	if m.IsSet("purge_scan_files") {
		api.PurgeScanFiles = *sonatypeiq.NewNullableString(m.PurgeScanFiles.ValueStringPointer())
	}
	if m.IsSet("session_timeout") {
		api.SessionTimeout = *sonatypeiq.NewNullableInt32(m.SessionTimeout.ValueInt32Pointer())
	}
	if m.IsSet("success_metrics_stage_id") {
		api.SuccessMetricsStageId = *sonatypeiq.NewNullableString(m.SuccessMetricsStageID.ValueStringPointer())
	}
	if m.IsSet("user_agent_suffix") {
		api.UserAgentSuffix = *sonatypeiq.NewNullableString(m.UserAgentSuffix.ValueStringPointer())
	}
	if m.IsSet("waived_component_upgrade_monitoring_enabled") {
		api.WaivedComponentUpgradeMonitoringEnabled = *sonatypeiq.NewNullableBool(m.WaivedComponentUpgradeMonitoringEnabled.ValueBoolPointer())
	}
	if m.IsSet("webhook_secret_passphrase") {
		api.WebhookSecretPassphrase = *sonatypeiq.NewNullableString(m.WebhookSecretPassphrase.ValueStringPointer())
	}
	return api
}

// MapFromApi updates the state from the API. Optional properties that are not managed in Terraform stay
// null, so that Sonatype IQ Server's own defaults never show up as drift. The webhook secret is never
// read back: Terraform keeps the value from the configuration.
func (m *SystemConfigResource) MapFromApi(api *sonatypeiq.SystemConfig) {
	prior := *m
	var fresh SystemConfigModel
	fresh.MapFromApi(api)

	m.ID = types.StringValue(common.STATE_ID_SYSTEM_CONFIGURATION)
	m.BaseURL = fresh.BaseURL
	m.ForceBaseURL = fresh.ForceBaseURL

	keep := func(attribute string) bool { return prior.IsSet(attribute) }
	if keep("advanced_reporting_insights_enabled") {
		m.AdvancedReportingInsightsEnabled = fresh.AdvancedReportingInsightsEnabled
	}
	if keep("advanced_search_csv_export_delimiter") {
		m.AdvancedSearchCsvExportDelimiter = fresh.AdvancedSearchCsvExportDelimiter
	}
	if keep("alp_observed_license_detection_enabled") {
		m.AlpObservedLicenseDetectionEnabled = fresh.AlpObservedLicenseDetectionEnabled
	}
	if keep("api_access_allow_list") {
		m.ApiAccessAllowList = fresh.ApiAccessAllowList
	}
	if keep("csrf_protection") {
		m.CsrfProtection = fresh.CsrfProtection
	}
	if keep("event_bus_max_thread_pool_size") {
		m.EventBusMaxThreadPoolSize = fresh.EventBusMaxThreadPoolSize
	}
	if keep("frame_ancestors_allowlist") {
		m.FrameAncestorsAllowlist = fresh.FrameAncestorsAllowlist
	}
	if keep("hds_url") {
		m.HdsURL = fresh.HdsURL
	}
	if keep("integrations_supported_version_count") {
		m.IntegrationsSupportedVersionCount = fresh.IntegrationsSupportedVersionCount
	}
	if keep("max_advanced_search_clause_count") {
		m.MaxAdvancedSearchClauseCount = fresh.MaxAdvancedSearchClauseCount
	}
	if keep("policy_monitoring_hour") {
		m.PolicyMonitoringHour = fresh.PolicyMonitoringHour
	}
	if keep("purge_scan_files") {
		m.PurgeScanFiles = fresh.PurgeScanFiles
	}
	if keep("session_timeout") {
		m.SessionTimeout = fresh.SessionTimeout
	}
	if keep("success_metrics_stage_id") {
		m.SuccessMetricsStageID = fresh.SuccessMetricsStageID
	}
	if keep("user_agent_suffix") {
		m.UserAgentSuffix = fresh.UserAgentSuffix
	}
	if keep("waived_component_upgrade_monitoring_enabled") {
		m.WaivedComponentUpgradeMonitoringEnabled = fresh.WaivedComponentUpgradeMonitoringEnabled
	}
}

// stringListValue maps a list of strings to a Terraform list, null when the API returned none.
func stringListValue(values []string) types.List {
	if values == nil {
		return types.ListNull(types.StringType)
	}
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, elements)
}

// stringsFromList maps a Terraform list of strings to a Go slice.
func stringsFromList(list types.List) []string {
	values := make([]string, 0, len(list.Elements()))
	for _, e := range list.Elements() {
		if s, ok := e.(types.String); ok {
			values = append(values, s.ValueString())
		}
	}
	return values
}
