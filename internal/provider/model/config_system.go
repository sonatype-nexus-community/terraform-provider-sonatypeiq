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
	api.BaseUrl = nullableString(m.BaseURL)
	api.ForceBaseUrl = nullableBool(m.ForceBaseURL)
	api.ADVANCED_REPORTING_INSIGHTS_ENABLED = nullableBool(m.AdvancedReportingInsightsEnabled)
	api.AdvancedSearchCSVExportDelimiter = nullableString(m.AdvancedSearchCsvExportDelimiter)
	api.AlpObservedLicenseDetectionEnabled = nullableBool(m.AlpObservedLicenseDetectionEnabled)
	api.ApiAccessAllowList = stringsFromList(m.ApiAccessAllowList)
	api.CsrfProtection = nullableBool(m.CsrfProtection)
	api.EventBusMaxThreadPoolSize = nullableInt32(m.EventBusMaxThreadPoolSize)
	api.FrameAncestorsAllowlist = stringsFromList(m.FrameAncestorsAllowlist)
	api.HdsUrl = nullableString(m.HdsURL)
	api.IntegrationsSupportedVersionCount = nullableInt32(m.IntegrationsSupportedVersionCount)
	api.MaxAdvancedSearchClauseCount = nullableInt32(m.MaxAdvancedSearchClauseCount)
	api.PolicyMonitoringHour = nullableInt32(m.PolicyMonitoringHour)
	api.PurgeScanFiles = nullableString(m.PurgeScanFiles)
	api.SessionTimeout = nullableInt32(m.SessionTimeout)
	api.SuccessMetricsStageId = nullableString(m.SuccessMetricsStageID)
	api.UserAgentSuffix = nullableString(m.UserAgentSuffix)
	api.WaivedComponentUpgradeMonitoringEnabled = nullableBool(m.WaivedComponentUpgradeMonitoringEnabled)
	api.WebhookSecretPassphrase = nullableString(m.WebhookSecretPassphrase)
	return api
}

// nullableString maps a Terraform string to the API, leaving it unset (not sent) when it is null.
func nullableString(v types.String) sonatypeiq.NullableString {
	if v.IsNull() {
		return sonatypeiq.NullableString{}
	}
	return *sonatypeiq.NewNullableString(v.ValueStringPointer())
}

// nullableBool maps a Terraform bool to the API, leaving it unset (not sent) when it is null.
func nullableBool(v types.Bool) sonatypeiq.NullableBool {
	if v.IsNull() {
		return sonatypeiq.NullableBool{}
	}
	return *sonatypeiq.NewNullableBool(v.ValueBoolPointer())
}

// nullableInt32 maps a Terraform int32 to the API, leaving it unset (not sent) when it is null.
func nullableInt32(v types.Int32) sonatypeiq.NullableInt32 {
	if v.IsNull() {
		return sonatypeiq.NullableInt32{}
	}
	return *sonatypeiq.NewNullableInt32(v.ValueInt32Pointer())
}

// MapFromApi updates the state from the API. Optional properties that are not managed in Terraform stay
// null, so that Sonatype IQ Server's own defaults never show up as drift. The webhook secret is never
// read back: Terraform keeps the value from the configuration.
func (m *SystemConfigResource) MapFromApi(api *sonatypeiq.SystemConfig) {
	var fresh SystemConfigModel
	fresh.MapFromApi(api)

	m.ID = types.StringValue(common.STATE_ID_SYSTEM_CONFIGURATION)
	m.BaseURL = fresh.BaseURL
	m.ForceBaseURL = fresh.ForceBaseURL
	m.AdvancedReportingInsightsEnabled = refreshIfManaged(m.AdvancedReportingInsightsEnabled, fresh.AdvancedReportingInsightsEnabled)
	m.AdvancedSearchCsvExportDelimiter = refreshIfManaged(m.AdvancedSearchCsvExportDelimiter, fresh.AdvancedSearchCsvExportDelimiter)
	m.AlpObservedLicenseDetectionEnabled = refreshIfManaged(m.AlpObservedLicenseDetectionEnabled, fresh.AlpObservedLicenseDetectionEnabled)
	m.ApiAccessAllowList = refreshIfManaged(m.ApiAccessAllowList, fresh.ApiAccessAllowList)
	m.CsrfProtection = refreshIfManaged(m.CsrfProtection, fresh.CsrfProtection)
	m.EventBusMaxThreadPoolSize = refreshIfManaged(m.EventBusMaxThreadPoolSize, fresh.EventBusMaxThreadPoolSize)
	m.FrameAncestorsAllowlist = refreshIfManaged(m.FrameAncestorsAllowlist, fresh.FrameAncestorsAllowlist)
	m.HdsURL = refreshIfManaged(m.HdsURL, fresh.HdsURL)
	m.IntegrationsSupportedVersionCount = refreshIfManaged(m.IntegrationsSupportedVersionCount, fresh.IntegrationsSupportedVersionCount)
	m.MaxAdvancedSearchClauseCount = refreshIfManaged(m.MaxAdvancedSearchClauseCount, fresh.MaxAdvancedSearchClauseCount)
	m.PolicyMonitoringHour = refreshIfManaged(m.PolicyMonitoringHour, fresh.PolicyMonitoringHour)
	m.PurgeScanFiles = refreshIfManaged(m.PurgeScanFiles, fresh.PurgeScanFiles)
	m.SessionTimeout = refreshIfManaged(m.SessionTimeout, fresh.SessionTimeout)
	m.SuccessMetricsStageID = refreshIfManaged(m.SuccessMetricsStageID, fresh.SuccessMetricsStageID)
	m.UserAgentSuffix = refreshIfManaged(m.UserAgentSuffix, fresh.UserAgentSuffix)
	m.WaivedComponentUpgradeMonitoringEnabled = refreshIfManaged(m.WaivedComponentUpgradeMonitoringEnabled, fresh.WaivedComponentUpgradeMonitoringEnabled)
}

// refreshIfManaged returns the value read from the API when the property is managed in Terraform (the
// prior value is not null), and keeps it null otherwise.
func refreshIfManaged[T attr.Value](prior, fresh T) T {
	if prior.IsNull() {
		return prior
	}
	return fresh
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
