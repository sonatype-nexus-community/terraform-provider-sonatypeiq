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
	"encoding/json"
	"fmt"
	"net/http"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"time"

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

// configJiraResource is the resource implementation.
type configJiraResource struct {
	common.BaseResource
}

// NewConfigJiraResource is a helper function to simplify the provider implementation.
func NewConfigJiraResource() resource.Resource {
	return &configJiraResource{}
}

// Metadata returns the resource type name.
func (r *configJiraResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_jira"
}

// Schema defines the schema for the resource.
func (r *configJiraResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Jira integration configuration for IQ Server",
		Attributes: map[string]schema.Attribute{
			"id":  sharedrschema.ResourceComputedString("Internal ID for Terraform State"),
			"url": sharedrschema.ResourceRequiredString("URL of the Jira server"),
			"username": func() schema.StringAttribute {
				attr := sharedrschema.ResourceOptionalString("Username for Jira authentication")
				attr.Validators = []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("password"),
					}...),
				}
				return attr
			}(),
			"password": func() schema.StringAttribute {
				attr := sharedrschema.ResourceSensitiveString("Password for Jira authentication")
				attr.Validators = []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("username"),
					}...),
				}
				return attr
			}(),
			"custom_fields": func() schema.StringAttribute {
				attr := sharedrschema.ResourceOptionalString("Custom fields for Jira integration as JSON string")
				attr.Validators = []validator.String{
					jsonObjectValidator{},
				}
				return attr
			}(),
			"last_updated": sharedrschema.ResourceLastUpdated(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *configJiraResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.ConfigJiraModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
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
func (r *configJiraResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.ConfigJiraModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	apiResponse := r.doRead(ctx, &resp.State, &resp.Diagnostics)
	if apiResponse == nil {
		return
	}

	// Update State based on Response
	if err := state.MapFromApi(apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error mapping API response to state",
			fmt.Sprintf("Could not map Jira configuration response: %v", err),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *configJiraResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.ConfigJiraModel
	var state model.ConfigJiraModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Check if credentials or custom_fields were removed - if so, delete first
	// The API merges fields, so removing username/password or custom_fields requires DELETE first
	needDelete := (!state.Username.IsNull() && plan.Username.IsNull()) ||
		(!state.Password.IsNull() && plan.Password.IsNull()) ||
		(!state.CustomFields.IsNull() && plan.CustomFields.IsNull())

	if needDelete {
		httpResponse, err := r.Client.ConfigJiraAPI.DeleteConfiguration2(r.AuthContext(ctx)).Execute()
		if err != nil && (httpResponse == nil || httpResponse.StatusCode != http.StatusNoContent) {
			resp.Diagnostics.AddError(
				"Error deleting Jira configuration before update",
				fmt.Sprintf("Could not delete configuration: %v", err),
			)
			return
		}
	}

	// Save plan password before doUpsert (which calls doRead that doesn't return password)
	planPassword := plan.Password

	r.doUpsert(ctx, &plan, &resp.State, &resp.Diagnostics)

	// Restore password from plan (doUpsert doesn't modify it, but let's be explicit)
	plan.Password = planPassword

	// Update State
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *configJiraResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	httpResponse, err := r.Client.ConfigJiraAPI.DeleteConfiguration2(r.AuthContext(ctx)).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			common.ERR_JIRA_CONFIGURATION_DID_NOT_EXIST,
			fmt.Sprintf("%v", err),
		)
		return
	}
}

func (r *configJiraResource) doRead(ctx context.Context, respState *tfsdk.State, respDiags *diag.Diagnostics) *sonatypeiq.ApiJiraConfigurationDTO {
	apiResponse, httpResponse, err := r.Client.ConfigJiraAPI.GetConfiguration2(r.AuthContext(ctx)).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			respState.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"Jira configuration did not exist",
				&err,
				httpResponse,
				respDiags,
			)
		} else {
			errors.HandleAPIError(
				common.ERR_FAILED_READING_JIRA_CONFIGURATION,
				&err,
				httpResponse,
				respDiags,
			)
		}
		return nil
	}

	return apiResponse
}

func (r *configJiraResource) doUpsert(ctx context.Context, model *model.ConfigJiraModel, respState *tfsdk.State, respDiags *diag.Diagnostics) {
	apiModel, err := model.MapToApi()
	if err != nil {
		respDiags.AddError(
			"Error mapping Terraform configuration to API model",
			fmt.Sprintf("Could not map Jira configuration: %v", err),
		)
		return
	}

	httpResponse, err := r.Client.ConfigJiraAPI.SetConfiguration2(r.AuthContext(ctx)).ApiJiraConfigurationDTO(*apiModel).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating/updating Jira configuration",
			&err,
			httpResponse,
			respDiags,
		)
		return
	} else if httpResponse.StatusCode != http.StatusNoContent {
		errors.HandleAPIError(
			"Upsertion of Jira configuration was not successful",
			&err,
			httpResponse,
			respDiags,
		)
		return
	}

	apiResponse := r.doRead(ctx, respState, respDiags)
	if apiResponse == nil {
		return
	}

	// Map response to State
	if err := model.MapFromApi(apiResponse); err != nil {
		respDiags.AddError(
			"Error mapping API response to state",
			fmt.Sprintf("Could not map Jira configuration response: %v", err),
		)
		return
	}
	model.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}

func (r *configJiraResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// jsonObjectValidator checks that a string holds a JSON object of objects, as accepted for the custom fields.
type jsonObjectValidator struct{}

func (jsonObjectValidator) Description(_ context.Context) string {
	return "value must be a JSON object whose values are JSON objects"
}

func (v jsonObjectValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (jsonObjectValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var fields map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &fields); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid custom_fields",
			fmt.Sprintf("custom_fields must be a JSON object whose values are JSON objects: %v", err))
	}
}
