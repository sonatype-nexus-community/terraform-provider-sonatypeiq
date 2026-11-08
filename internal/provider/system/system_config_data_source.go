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

package system

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	"github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &systemConfigDataSource{}
	_ datasource.DataSourceWithConfigure = &systemConfigDataSource{}
)

// SystemConfigDataSource is a helper function to simplify the provider implementation.
func SystemConfigDataSource() datasource.DataSource {
	return &systemConfigDataSource{}
}

// systemConfigDataSource is the data source implementation.
type systemConfigDataSource struct {
	common.BaseDataSource
}

// Metadata returns the data source type name.
func (d *systemConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_config"
}

// Schema defines the schema for the data source.
func (d *systemConfigDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Description: "Use this data source to get System Configuration. Each property documents the Sonatype IQ Server release that introduced it; properties the connected Sonatype IQ Server does not support are returned as null. " +
			"See https://help.sonatype.com/en/configuration-rest-api.html\n\n",
		Attributes: map[string]tfschema.Attribute{
			"id":             schema.DataSourceComputedString("The ID of this resource."),
			"base_url":       schema.DataSourceComputedString("Base URL for Sonatype IQ Server"),
			"force_base_url": schema.DataSourceComputedBool("Should the Base URL be forced?"),
			"advanced_reporting_insights_enabled": tfschema.BoolAttribute{
				Computed:    true,
				Description: systemConfigDescription("advanced_reporting_insights_enabled", "Whether non-anonymized telemetry data is sent to Sonatype for Advanced Reporting Insights."),
			},
			"advanced_search_csv_export_delimiter": tfschema.StringAttribute{
				Computed:    true,
				Description: systemConfigDescription("advanced_search_csv_export_delimiter", "Delimiter used when exporting Advanced Search results to CSV."),
			},
			"alp_observed_license_detection_enabled": tfschema.BoolAttribute{
				Computed:    true,
				Description: systemConfigDescription("alp_observed_license_detection_enabled", "Whether observed license detection for additional ecosystems is enabled."),
			},
			"api_access_allow_list": tfschema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: systemConfigDescription("api_access_allow_list", "Usernames that are allowed to access the public REST API."),
			},
			"csrf_protection": tfschema.BoolAttribute{
				Computed:    true,
				Description: systemConfigDescription("csrf_protection", "Whether Cross-Site Request Forgery (CSRF) protection is enabled."),
			},
			"event_bus_max_thread_pool_size": tfschema.Int32Attribute{
				Computed:    true,
				Description: systemConfigDescription("event_bus_max_thread_pool_size", "Maximum size of the event bus thread pool."),
			},
			"frame_ancestors_allowlist": tfschema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: systemConfigDescription("frame_ancestors_allowlist", "Domain patterns that are allowed to embed Sonatype IQ Server in a frame. Sonatype IQ Server always includes `'self'`."),
			},
			"hds_url": tfschema.StringAttribute{
				Computed:    true,
				Description: systemConfigDescription("hds_url", "URL of the Sonatype Hosted Data Services (HDS)."),
			},
			"integrations_supported_version_count": tfschema.Int32Attribute{
				Computed:    true,
				Description: systemConfigDescription("integrations_supported_version_count", "How many of the most recent Sonatype integration versions are permitted to perform policy evaluations."),
			},
			"max_advanced_search_clause_count": tfschema.Int32Attribute{
				Computed:    true,
				Description: systemConfigDescription("max_advanced_search_clause_count", "Maximum number of clauses (query parameters) in an Advanced Search."),
			},
			"policy_monitoring_hour": tfschema.Int32Attribute{
				Computed:    true,
				Description: systemConfigDescription("policy_monitoring_hour", "Hour of the day (0-23, system time) at which Continuous Monitoring starts."),
			},
			"purge_scan_files": tfschema.StringAttribute{
				Computed:    true,
				Description: systemConfigDescription("purge_scan_files", "Whether scan files are purged after evaluation (`withReports`) or not (null)."),
			},
			"session_timeout": tfschema.Int32Attribute{
				Computed:    true,
				Description: systemConfigDescription("session_timeout", "User session timeout, in minutes."),
			},
			"success_metrics_stage_id": tfschema.StringAttribute{
				Computed:    true,
				Description: systemConfigDescription("success_metrics_stage_id", "The Lifecycle stage that Success Metrics is restricted to."),
			},
			"user_agent_suffix": tfschema.StringAttribute{
				Computed:    true,
				Description: systemConfigDescription("user_agent_suffix", "Suffix appended to the User-Agent that Sonatype IQ Server sends."),
			},
			"waived_component_upgrade_monitoring_enabled": tfschema.BoolAttribute{
				Computed:    true,
				Description: systemConfigDescription("waived_component_upgrade_monitoring_enabled", "Whether waived components are monitored for available upgrades."),
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *systemConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.SystemConfigModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Only ask for the properties this Sonatype IQ Server release is documented to support
	apiResponse, httpResponse, err := d.Client.ConfigurationAPI.GetConfiguration1(d.AuthContext(ctx)).Property(model.RequestedPropertiesFor(d.IqVersion)).Execute()

	if err != nil && httpResponse != nil && httpResponse.StatusCode == http.StatusBadRequest {
		// The server did not recognise one of the property names: fall back to the properties every release has,
		// rather than breaking configurations that only ever needed those.
		tflog.Warn(ctx, fmt.Sprintf("Sonatype IQ Server %v rejected the requested System Configuration properties - reading only the base properties", d.IqVersion))
		resp.Diagnostics.AddWarning(
			"Some System Configuration properties could not be read",
			fmt.Sprintf("Sonatype IQ Server release %d rejected the request for the full set of System Configuration properties, so only `base_url` and `force_base_url` have been read. All other properties are null.", d.IqVersion),
		)
		apiResponse, httpResponse, err = d.Client.ConfigurationAPI.GetConfiguration1(d.AuthContext(ctx)).Property(model.BaseProperties()).Execute()
	}

	if err != nil {
		errors.HandleAPIError(
			common.ERR_FAILED_READING_SYSTEM_CONFIG,
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.AddAPIErrorDiagnostic(&resp.Diagnostics, "read", "System Configuration", httpResponse, err)
		return
	}

	// Assign Response Data to State
	data.ID = types.StringValue("system-configuration")
	data.MapFromApi(apiResponse)

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
