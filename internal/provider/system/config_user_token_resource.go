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
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
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

// userTokenConfigurationIntroducedIn is the Sonatype IQ Server release that introduced the User Token configuration endpoint.
const userTokenConfigurationIntroducedIn int32 = 199

// configUserTokenResource is the resource implementation.
type configUserTokenResource struct {
	common.BaseResource
}

// NewConfigUserTokenResource is a helper function to simplify the provider implementation.
func NewConfigUserTokenResource() resource.Resource {
	return &configUserTokenResource{}
}

// Metadata returns the resource type name.
func (r *configUserTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_token_configuration"
}

// Schema defines the schema for the resource.
func (r *configUserTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage User Token configuration for IQ Server. Requires Sonatype IQ Server release 199 or later.",
		Attributes: map[string]schema.Attribute{
			"id": sharedrschema.ResourceComputedString("Internal ID for Terraform State"),
			"user_token_default_expiration_days": schema.Int32Attribute{
				Description: "Default expiration time for user tokens in days (1-365). Requires Sonatype IQ Server release 199 or later.",
				Required:    true,
				Validators: []validator.Int32{
					int32validator.Between(1, 365),
				},
			},
			"last_updated": sharedrschema.ResourceLastUpdated(),
		},
	}
}

// ModifyPlan validates that the server supports this resource.
func (r *configUserTokenResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip when destroying (plan is null)
	if req.Plan.Raw.IsNull() {
		return
	}

	// Check server version - requires release 199 or later
	common.RequireIqRelease(&resp.Diagnostics, r.IqVersion, userTokenConfigurationIntroducedIn, "user_token_configuration resource")
}

// Create creates the resource and sets the initial Terraform state.
func (r *configUserTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.ConfigUserTokenModel
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
func (r *configUserTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.ConfigUserTokenModel
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
	state.MapFromApi(apiResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *configUserTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.ConfigUserTokenModel
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

// Delete deletes the resource and removes the Terraform state on success.
func (r *configUserTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// DELETE resets the configuration to defaults
	// We need to specify the property to reset
	apiResponse, httpResponse, err := r.Client.UserTokenConfigurationAPI.ResetConfiguration(r.AuthContext(ctx)).
		Property([]string{"userTokenDefaultExpirationDays"}).
		Execute()

	if err != nil || httpResponse.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			common.ERR_USER_TOKEN_CONFIGURATION_DID_NOT_EXIST,
			fmt.Sprintf("%v", err),
		)
		return
	}

	// Log the response for debugging
	tflog.Debug(ctx, fmt.Sprintf("Reset user token configuration: %v", apiResponse))
}

func (r *configUserTokenResource) doRead(ctx context.Context, respState *tfsdk.State, respDiags *diag.Diagnostics) *sonatypeiq.ApiUserTokenConfigurationDTO {
	apiResponse, httpResponse, err := r.Client.UserTokenConfigurationAPI.GetConfiguration8(r.AuthContext(ctx)).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			respState.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"User Token configuration did not exist",
				&err,
				httpResponse,
				respDiags,
			)
		} else {
			errors.HandleAPIError(
				common.ERR_FAILED_READING_USER_TOKEN_CONFIGURATION,
				&err,
				httpResponse,
				respDiags,
			)
		}
		return nil
	}

	return apiResponse
}

func (r *configUserTokenResource) doUpsert(ctx context.Context, model *model.ConfigUserTokenModel, respState *tfsdk.State, respDiags *diag.Diagnostics) {
	_, httpResponse, err := r.Client.UserTokenConfigurationAPI.UpdateConfiguration(r.AuthContext(ctx)).
		ApiUserTokenConfigurationDTO(*model.MapToApi()).
		Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating/updating User Token configuration",
			&err,
			httpResponse,
			respDiags,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.HandleAPIError(
			"Upsertion of User Token configuration was not successful",
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
	model.MapFromApi(apiResponse)
	model.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}

func (r *configUserTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
