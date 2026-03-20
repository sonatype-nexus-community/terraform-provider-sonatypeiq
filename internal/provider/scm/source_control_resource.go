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

package scm

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

type sourceControlResource struct {
	common.BaseResource
}

// NewSourceControlResource is a helper function to simplify the provider implementation.
func NewSourceControlResource() resource.Resource {
	return &sourceControlResource{}
}

// Metadata returns the resource type name.
func (r *sourceControlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source_control"
}

// Schema defines the provider inputs.
func (r *sourceControlResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":       sharedrschema.ResourceComputedString("Internal ID for Terraform State"),
			"owner_id": sharedrschema.ResourceRequiredString("Must be a valid organization or application ID, for the root organization use `ROOT_ORGANIZATION_ID`"),
			"owner_type": sharedrschema.ResourceRequiredStringEnum(
				"The type of the owner, must be one of 'organization' or 'application'.",
				common.OWNER_TYPE_APPLICATION,
				common.OWNER_TYPE_ORGANIZATION,
			),
			"repository_url":                    sharedrschema.ResourceOptionalString("The SCM provider URL for the repository, only valid for `owner_type` of `application`"),
			"base_branch":                       sharedrschema.ResourceOptionalString("The default branch to use."),
			"pull_request_commenting_enabled":   sharedrschema.ResourceOptionalBool("Set to true to enable the Pull Request Commenting feature."),
			"source_control_evaluation_enabled": sharedrschema.ResourceOptionalBool("Set to true to enable Sonatype Lifecycle triggered source control evaluations."),
			"user_name": sharedrschema.ResourceOptionalStringWithValidators(
				"The user name to use when setting `scm_provider` to `bitbucket`.",
				stringvalidator.AlsoRequires(path.MatchRoot("scm_provider")),
			),
			"remediation_pull_requests_enabled": sharedrschema.ResourceComputedOptionalBool("Set to true to enable the Automated Pull Requests feature."),
			"scm_provider": sharedrschema.ResourceOptionalStringEnum(
				"The type of SCM Provider, must be one of 'azure, bitbucket, github or gitlab'. This is required when the organization is set to `ROOT_ORGANIZATION_ID`",
				common.SCM_PROVIDER_AZURE_DEVOPS,
				common.SCM_PROVIDER_BITBUCKET,
				common.SCM_PROVIDER_GITHUB,
				common.SCM_PROVIDER_GITLAB,
			),
			"token": func() schema.StringAttribute {
				attr := sharedrschema.ResourceOptionalStringWithPlanModifier(
					"The token for use with the SCM Provider",
					stringplanmodifier.UseStateForUnknown(),
				)
				attr.Computed = false
				attr.Validators = append(attr.Validators, stringvalidator.AlsoRequires(path.MatchRoot("scm_provider")))
				return attr
			}(),
			"last_updated": sharedrschema.ResourceLastUpdated(),
		},
	}
}

// ModifyPlan implements resource.ResourceWithModifyPlan.
func (r *sourceControlResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Nothing to do on destroy
	if req.Plan.Raw.IsNull() {
		return
	}

	var remediationPRs types.Bool
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("remediation_pull_requests_enabled"), &remediationPRs)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only set a default if the user hasn't explicitly provided a value
	if !remediationPRs.IsNull() {
		return
	}

	var ownerID types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("owner_id"), &ownerID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Inject ROOT ORG level defaults if they are not supplied in the Plan
	if ownerID.ValueString() == common.ROOT_ORGANIZATION_ID {
		resp.Plan.SetAttribute(ctx, path.Root("remediation_pull_requests_enabled"), types.BoolValue(true))
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *sourceControlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.SourceControlModelResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := r.Client.SourceControlAPI.AddSourceControl(
		r.AuthContext(ctx),
		strings.Trim(plan.OwnerType.ValueString(), "\""),
		strings.Trim(plan.OwnerID.ValueString(), "\""),
	).ApiSourceControlDTO(*plan.MapToApi()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating Source Control configuration",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.HandleAPIError(
			"Creation of Source Control configuration was not successful",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// Map response to State
	plan.MapFromApi(apiResponse)

	// Update State
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *sourceControlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.SourceControlModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := r.Client.SourceControlAPI.GetSourceControl1(
		r.AuthContext(ctx),
		state.OwnerType.ValueString(),
		state.OwnerID.ValueString(),
	).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"Source Control configuraiton does not exist for OwnerType/ID combination",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		} else {
			errors.HandleAPIError(
				common.ERR_FAILED_READING_SCM_CONFIGURATION,
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		}
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
func (r *sourceControlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.SourceControlModelResource
	var state model.SourceControlModelResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := r.Client.SourceControlAPI.UpdateSourceControl(
		r.AuthContext(ctx),
		state.OwnerType.ValueString(),
		state.OwnerID.ValueString(),
	).ApiSourceControlDTO(*plan.MapToApi()).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error updating Source Control configuration",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.HandleAPIError(
			"Updating Source Control configuration was not successful",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// Map response to State
	plan.MapFromApi(apiResponse)
	plan.Token = state.Token

	// Update State
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sourceControlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state model.SourceControlModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	httpResponse, err := r.Client.SourceControlAPI.DeleteSourceControl(
		r.AuthContext(ctx),
		state.OwnerType.ValueString(),
		state.OwnerID.ValueString(),
	).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			fmt.Sprintf(common.ERR_SOURCE_CONTROL_CONFIGURATION_DID_NOT_EXIST, state.ID.ValueString()),
			fmt.Sprintf("%v", err),
		)
		return
	}
}

func (r *sourceControlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: <owner-type>,<owner-id>. Got: %q", req.ID),
		)
		return
	}
	if idParts[0] != common.OWNER_TYPE_APPLICATION && idParts[0] != common.OWNER_TYPE_ORGANIZATION {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier prefix",
			fmt.Sprintf("Expected import identifier to start with '%s' or '%s'. Got: %q", common.OWNER_TYPE_APPLICATION, common.OWNER_TYPE_ORGANIZATION, idParts[0]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), fmt.Sprintf(common.SCM_CONFIG_ID_FORMAT, idParts[0], idParts[1]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner_type"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner_id"), idParts[1])...)
}
