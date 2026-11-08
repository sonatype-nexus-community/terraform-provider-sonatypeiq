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
	"fmt"
	"strings"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// SystemConfigPropertyInfo describes a System Configuration property and when it became available.
type SystemConfigPropertyInfo struct {
	// Attribute is the Terraform attribute name
	Attribute string
	// ApiProperty is the property name used by the Configuration REST API
	ApiProperty sonatypeiq.SystemConfigProperty
	// IntroducedIn is the Sonatype IQ Server release (e.g. 175) that introduced the property.
	// 0 means that no release is known, so the property is assumed to be available in every release
	// supported by this provider.
	IntroducedIn int32
}

// SystemConfigProperties lists every property managed by sonatypeiq_system_config.
//
// IntroducedIn is the first Sonatype IQ Server release in which the property can be set through the
// Configuration REST API. It was established from the Sonatype IQ Server release history: the release
// that first shipped the property in the REST API's list of supported properties. Where Sonatype's
// published documentation (https://help.sonatype.com/en/configuration-rest-api.html) also states a release
// it agrees, except for alp_observed_license_detection_enabled (the product shipped it in 161, with a
// database schema correction in 162, but the documentation says 165).
//
// A value of 0 would mean that no release is known, in which case the property is assumed to be
// available in every release supported by this provider. Add new properties here with their release and
// the gating, the schema descriptions and the generated documentation all follow.
var SystemConfigProperties = []SystemConfigPropertyInfo{
	{Attribute: "base_url", ApiProperty: sonatypeiq.BASE_URL, IntroducedIn: 138},
	{Attribute: "force_base_url", ApiProperty: sonatypeiq.FORCE_BASE_URL, IntroducedIn: 138},
	{Attribute: "advanced_reporting_insights_enabled", ApiProperty: sonatypeiq.ADVANCED_REPORTING_INSIGHTS_ENABLED, IntroducedIn: 175},
	{Attribute: "advanced_search_csv_export_delimiter", ApiProperty: sonatypeiq.ADVANCED_SEARCH_CSV_EXPORT_DELIMITER, IntroducedIn: 142},
	{Attribute: "alp_observed_license_detection_enabled", ApiProperty: sonatypeiq.ALP_OBSERVED_LICENSE_DETECTION_ENABLED, IntroducedIn: 161},
	{Attribute: "api_access_allow_list", ApiProperty: sonatypeiq.API_ACCESS_ALLOW_LIST, IntroducedIn: 175},
	{Attribute: "csrf_protection", ApiProperty: sonatypeiq.CSRF_PROTECTION, IntroducedIn: 142},
	{Attribute: "event_bus_max_thread_pool_size", ApiProperty: sonatypeiq.EVENT_BUS_MAX_THREAD_POOL_SIZE, IntroducedIn: 142},
	{Attribute: "frame_ancestors_allowlist", ApiProperty: sonatypeiq.FRAME_ANCESTORS_ALLOWLIST, IntroducedIn: 142},
	{Attribute: "hds_url", ApiProperty: sonatypeiq.HDS_URL, IntroducedIn: 142},
	{Attribute: "integrations_supported_version_count", ApiProperty: sonatypeiq.INTEGRATIONS_SUPPORTED_VERSION_COUNT, IntroducedIn: 197},
	{Attribute: "max_advanced_search_clause_count", ApiProperty: sonatypeiq.MAX_ADVANCED_SEARCH_CLAUSE_COUNT, IntroducedIn: 142},
	{Attribute: "policy_monitoring_hour", ApiProperty: sonatypeiq.POLICY_MONITORING_HOUR, IntroducedIn: 142},
	{Attribute: "purge_scan_files", ApiProperty: sonatypeiq.PURGE_SCAN_FILES, IntroducedIn: 152},
	{Attribute: "session_timeout", ApiProperty: sonatypeiq.SESSION_TIMEOUT, IntroducedIn: 143},
	{Attribute: "success_metrics_stage_id", ApiProperty: sonatypeiq.SUCCESS_METRICS_STAGE_ID, IntroducedIn: 182},
	{Attribute: "user_agent_suffix", ApiProperty: sonatypeiq.USER_AGENT_SUFFIX, IntroducedIn: 142},
	{Attribute: "waived_component_upgrade_monitoring_enabled", ApiProperty: sonatypeiq.WAIVED_COMPONENT_UPGRADE_MONITORING_ENABLED, IntroducedIn: 162},
	{Attribute: "webhook_secret_passphrase", ApiProperty: sonatypeiq.WEBHOOK_SECRET_PASSPHRASE, IntroducedIn: 142},
}

// SystemConfigPropertyByAttribute returns the information for a Terraform attribute.
func SystemConfigPropertyByAttribute(attribute string) (SystemConfigPropertyInfo, bool) {
	for _, p := range SystemConfigProperties {
		if p.Attribute == attribute {
			return p, true
		}
	}
	return SystemConfigPropertyInfo{}, false
}

// IsSupportedBy reports whether a Sonatype IQ Server of the given release supports this property.
// An undetected server version (0) is never treated as unsupported.
func (p SystemConfigPropertyInfo) IsSupportedBy(iqVersion int32) bool {
	return iqVersion == 0 || p.IntroducedIn <= iqVersion
}

// AvailabilityNote is appended to attribute descriptions so users can see when a property was introduced.
func (p SystemConfigPropertyInfo) AvailabilityNote() string {
	if p.IntroducedIn == 0 {
		return " The release that introduced this property is not known."
	}
	return fmt.Sprintf(" Requires Sonatype IQ Server release %d or later.", p.IntroducedIn)
}

// UnsupportedMessage explains why a property cannot be used with the connected Sonatype IQ Server.
func (p SystemConfigPropertyInfo) UnsupportedMessage(iqVersion int32) string {
	return fmt.Sprintf(
		"`%s` was introduced in Sonatype IQ Server release %d, but the connected server is release %d. "+
			"Remove `%s` from your configuration, or upgrade Sonatype IQ Server to release %d or later.",
		p.Attribute, p.IntroducedIn, iqVersion, p.Attribute, p.IntroducedIn,
	)
}

// SystemConfigAvailabilityMarkdown renders when each property became available as a Markdown table.
// It is generated from SystemConfigProperties so the documentation always matches the version gating.
func SystemConfigAvailabilityMarkdown() string {
	var b strings.Builder
	b.WriteString("| Attribute | Configuration REST API property | Introduced in Sonatype IQ Server release |\n")
	b.WriteString("|---|---|---|\n")
	for _, p := range SystemConfigProperties {
		introduced := "not known"
		if p.IntroducedIn != 0 {
			introduced = fmt.Sprintf("%d", p.IntroducedIn)
		}
		fmt.Fprintf(&b, "| `%s` | `%s` | %s |\n", p.Attribute, p.ApiProperty, introduced)
	}
	return b.String()
}
