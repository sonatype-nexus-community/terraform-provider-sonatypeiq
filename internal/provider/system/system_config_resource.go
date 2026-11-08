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
	"regexp"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// systemConfigResource is the resource implementation.
type systemConfigResource struct {
	common.BaseResource
}

// NewSystemConfigResource is a helper function to simplify the provider implementation.
func NewSystemConfigResource() resource.Resource {
	return &systemConfigResource{}
}

// Metadata returns the resource type name.
func (r *systemConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_config"
}

// systemConfigDescription builds an attribute description that states when the property was introduced.
func systemConfigDescription(attribute, description string) string {
	info, ok := model.SystemConfigPropertyByAttribute(attribute)
	if !ok {
		return description
	}
	return description + info.AvailabilityNote()
}

// Schema defines the schema for the resource.
func (r *systemConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this resource to manage System Configuration. Every optional property is only managed when it is set: " +
			"properties you leave out keep Sonatype IQ Server's own value, and a property you remove from your configuration " +
			"is deleted so that Sonatype IQ Server reverts to its default. Each property documents the Sonatype IQ Server release " +
			"that introduced it - setting a property that the connected Sonatype IQ Server does not support fails at plan time. " +
			"See https://help.sonatype.com/en/configuration-rest-api.html\n\n",
		Attributes: map[string]schema.Attribute{
			"id":             sharedrschema.ResourceComputedString("Internal ID for Terraform State"),
			"base_url":       sharedrschema.ResourceRequiredString("Base URL for Sonatype IQ Server. See https://help.sonatype.com/en/configuration-rest-api.html#base-url--required-"),
			"force_base_url": sharedrschema.ResourceRequiredBool("Should the Base URL be forced? See https://help.sonatype.com/en/configuration-rest-api.html#force-the-base-url"),
			"advanced_reporting_insights_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: systemConfigDescription("advanced_reporting_insights_enabled", "Allow Sonatype IQ Server to send non-anonymized telemetry data to Sonatype for Advanced Reporting Insights. Sonatype IQ Server default: `true`."),
			},
			"advanced_search_csv_export_delimiter": schema.StringAttribute{
				Optional:    true,
				Description: systemConfigDescription("advanced_search_csv_export_delimiter", "Single character used as the delimiter when exporting Advanced Search results to CSV. Sonatype IQ Server default: `,`."),
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 1)},
			},
			"alp_observed_license_detection_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: systemConfigDescription("alp_observed_license_detection_enabled", "Enable observed license detection for additional ecosystems for users of Lifecycle with Advanced Legal Pack. Sonatype IQ Server default: `true`."),
			},
			"api_access_allow_list": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: systemConfigDescription("api_access_allow_list", "Usernames that are allowed to access the public REST API. Sonatype IQ Server removes duplicates and does not keep the order, so this is a set. To remove the restriction, remove the attribute. Sonatype IQ Server default: unset."),
				Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
			},
			"csrf_protection": schema.BoolAttribute{
				Optional:    true,
				Description: systemConfigDescription("csrf_protection", "Enable Cross-Site Request Forgery (CSRF) protection. Sonatype IQ Server default: `true`."),
			},
			"event_bus_max_thread_pool_size": schema.Int32Attribute{
				Optional:    true,
				Description: systemConfigDescription("event_bus_max_thread_pool_size", "Maximum size of the event bus thread pool. Sonatype IQ Server default: `500`."),
				Validators:  []validator.Int32{int32validator.AtLeast(1)},
			},
			"frame_ancestors_allowlist": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: systemConfigDescription("frame_ancestors_allowlist", "Domain patterns that are allowed to embed Sonatype IQ Server in a frame, e.g. `[\"*first.com\", \"second.*\"]`. Sonatype IQ Server always allows `'self'` itself, so you do not need to list it. Use `[\"'none'\"]` on its own to forbid all framing. To remove the allowlist, remove the attribute. Sonatype IQ Server default: unset."),
				Validators:  []validator.Set{setvalidator.SizeAtLeast(1), noneMustBeAlone{}},
			},
			"hds_url": schema.StringAttribute{
				Optional:    true,
				Description: systemConfigDescription("hds_url", "URL of the Sonatype Hosted Data Services (HDS). Sonatype IQ Server adds a trailing `/`. Sonatype IQ Server default: `https://clm.sonatype.com/`."),
			},
			"integrations_supported_version_count": schema.Int32Attribute{
				Optional:    true,
				Description: systemConfigDescription("integrations_supported_version_count", "How many of the most recent Sonatype integration versions are permitted to perform policy evaluations. Must be a positive number. Sonatype IQ Server default: unset."),
				Validators:  []validator.Int32{int32validator.AtLeast(1)},
			},
			"max_advanced_search_clause_count": schema.Int32Attribute{
				Optional:    true,
				Description: systemConfigDescription("max_advanced_search_clause_count", "Maximum number of clauses (query parameters) in an Advanced Search. Sonatype IQ Server default: `2048`."),
				Validators:  []validator.Int32{int32validator.AtLeast(1)},
			},
			"policy_monitoring_hour": schema.Int32Attribute{
				Optional:    true,
				Description: systemConfigDescription("policy_monitoring_hour", "Hour of the day (0-23, system time) at which Continuous Monitoring starts. Sonatype IQ Server default: `0`."),
				Validators:  []validator.Int32{int32validator.Between(0, 23)},
			},
			"purge_scan_files": schema.StringAttribute{
				Optional:    true,
				Description: systemConfigDescription("purge_scan_files", "Purge scan files after evaluation. The only supported value is `withReports`. Sonatype IQ Server default: unset."),
				Validators:  []validator.String{stringvalidator.OneOf("withReports")},
			},
			"session_timeout": schema.Int32Attribute{
				Optional:    true,
				Description: systemConfigDescription("session_timeout", "User session timeout, in minutes (3 to 120). Sonatype IQ Server default: `30`."),
				Validators:  []validator.Int32{int32validator.Between(3, 120)},
			},
			"success_metrics_stage_id": schema.StringAttribute{
				Optional:    true,
				Description: systemConfigDescription("success_metrics_stage_id", "Restrict Success Metrics to a single Lifecycle stage. One of `source`, `build`, `stage-release`, `release`, `operate`, `proxy`, `compliance`. Sonatype IQ Server default: unset."),
				Validators:  []validator.String{stringvalidator.OneOf("source", "build", "stage-release", "release", "operate", "proxy", "compliance")},
			},
			"user_agent_suffix": schema.StringAttribute{
				Optional:    true,
				Description: systemConfigDescription("user_agent_suffix", "Suffix appended to the User-Agent that Sonatype IQ Server sends: at most 128 characters, with no control characters. Sonatype IQ Server default: unset."),
				Validators: []validator.String{
					stringvalidator.LengthAtMost(128),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[^\x00-\x1f\x7f]*$`), "must not contain control characters"),
				},
			},
			"waived_component_upgrade_monitoring_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: systemConfigDescription("waived_component_upgrade_monitoring_enabled", "Monitor waived components for available upgrades. Sonatype IQ Server default: `false`."),
			},
			"webhook_secret_passphrase": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: systemConfigDescription("webhook_secret_passphrase", "Secret key passphrase used to sign webhook payloads. It is never read back from Sonatype IQ Server, so Terraform cannot detect it being changed outside of Terraform."),
			},
			"last_updated": sharedrschema.ResourceLastUpdated(),
		},
	}
}

// ModifyPlan fails the plan when a property is set that the connected Sonatype IQ Server is too old to support.
func (r *systemConfigResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		// Destroying
		return
	}

	var plan model.SystemConfigResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.addUnsupportedPropertyErrors(&plan, &resp.Diagnostics)
}

// addUnsupportedPropertyErrors adds an error for every property that the connected Sonatype IQ Server does not support.
// It does nothing when the server version is unknown (e.g. while validating).
func (r *systemConfigResource) addUnsupportedPropertyErrors(plan *model.SystemConfigResource, diags *diag.Diagnostics) {
	for _, p := range plan.UnsupportedProperties(r.IqVersion) {
		diags.AddAttributeError(
			path.Root(p.Attribute),
			"Property is not supported by this Sonatype IQ Server version",
			p.UnsupportedMessage(r.IqVersion),
		)
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *systemConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.SystemConfigResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
		return
	}

	r.doUpsert(ctx, &plan, &resp.State, &resp.Diagnostics)

	// Update State
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *systemConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.SystemConfigResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	apiResponse := r.doRead(ctx, &resp.State, &resp.Diagnostics, state.ApiPropertiesInUse())
	if apiResponse == nil {
		return
	}

	// Update State based on Response
	state.MapFromApi(apiResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *systemConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state model.SystemConfigResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
		return
	}

	// Properties that are no longer set in Terraform are deleted, so Sonatype IQ Server reverts to their default
	if removed := state.RemovedProperties(&plan); len(removed) > 0 {
		removedApi := make([]sonatypeiq.SystemConfigProperty, 0, len(removed))
		for _, p := range removed {
			removedApi = append(removedApi, p.ApiProperty)
		}
		httpResponse, err := r.Client.ConfigurationAPI.DeleteConfiguration1(r.AuthContext(ctx)).Property(removedApi).Execute()
		if err != nil || httpResponse.StatusCode != http.StatusNoContent {
			errors.HandleAPIError(
				"Error removing System Property configuration",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
			return
		}
	}

	r.doUpsert(ctx, &plan, &resp.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update State
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
// Only the properties managed by this resource are deleted.
func (r *systemConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.SystemConfigResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	httpResponse, err := r.Client.ConfigurationAPI.DeleteConfiguration1(r.AuthContext(ctx)).Property(managedProperties(state.ApiPropertiesInUse())).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			common.ERR_SYSTEM_CONFIGURATION_DID_NOT_EXIST,
			fmt.Sprintf("%v", err),
		)
		return
	}
}

// managedProperties returns the base properties plus the given properties, without duplicates.
func managedProperties(props []sonatypeiq.SystemConfigProperty) []sonatypeiq.SystemConfigProperty {
	all := model.BaseProperties()
	seen := map[sonatypeiq.SystemConfigProperty]bool{}
	for _, p := range all {
		seen[p] = true
	}
	for _, p := range props {
		if !seen[p] {
			seen[p] = true
			all = append(all, p)
		}
	}
	return all
}

func (r *systemConfigResource) doRead(ctx context.Context, respState *tfsdk.State, respDiags *diag.Diagnostics, props []sonatypeiq.SystemConfigProperty) *sonatypeiq.SystemConfig {
	apiResponse, httpResponse, err := r.Client.ConfigurationAPI.GetConfiguration1(r.AuthContext(ctx)).Property(managedProperties(props)).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			respState.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"System configuration did not exist",
				&err,
				httpResponse,
				respDiags,
			)
		} else {
			errors.HandleAPIError(
				common.ERR_FAILED_READING_SYSTEM_CONFIG,
				&err,
				httpResponse,
				respDiags,
			)
		}
		return nil
	}

	return apiResponse
}

func (r *systemConfigResource) doUpsert(ctx context.Context, model *model.SystemConfigResource, respState *tfsdk.State, respDiags *diag.Diagnostics) {
	// Refuse properties that this Sonatype IQ Server is too old to support, rather than send them
	r.addUnsupportedPropertyErrors(model, respDiags)
	if respDiags.HasError() {
		return
	}

	httpResponse, err := r.Client.ConfigurationAPI.SetConfiguration1(r.AuthContext(ctx)).SystemConfig(*model.MapToApi()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating/updating System Property configuration",
			&err,
			httpResponse,
			respDiags,
		)
		return
	} else if httpResponse.StatusCode != http.StatusNoContent {
		errors.HandleAPIError(
			"Upsertion of System Property configuration was not successful",
			&err,
			httpResponse,
			respDiags,
		)
		return
	}

	apiResponse := r.doRead(ctx, respState, respDiags, model.ApiPropertiesInUse())
	if apiResponse == nil {
		return
	}

	// Map response to State
	model.MapFromApi(apiResponse)
	model.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}

func (r *systemConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
