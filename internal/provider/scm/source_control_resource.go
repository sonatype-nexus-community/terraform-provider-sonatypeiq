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
	"io"
	"strings"
	"terraform-provider-sonatypeiq/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

var _ resource.ResourceWithImportState = &sourceControlResource{}

type sourceControlResource struct {
	common.BaseResource
}

type sourceControlModelResource struct {
	ID                              types.String `tfsdk:"owner_id"`
	OwnerType                       types.String `tfsdk:"owner_type"`
	RepositoryUrl                   types.String `tfsdk:"repository_url"`
	BaseBranch                      types.String `tfsdk:"base_branch"`
	UserName                        types.String `tfsdk:"user_name"`
	RemediationPullRequestsEnabled  types.Bool   `tfsdk:"remediation_pull_requests_enabled"`
	PullRequestCommentingEnabled    types.Bool   `tfsdk:"pull_request_commenting_enabled"`
	SourceControlEvaluationsEnabled types.Bool   `tfsdk:"source_control_evaluation_enabled"`
	Token                           types.String `tfsdk:"token"`
	ScmProvider                     types.String `tfsdk:"scm_provider"` // This is provider in the rest API but provider is a reserved keyword
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
			"owner_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Must be a valid organization or application ID, for the root organization use `ROOT_ORGANIZATION_ID`",
			},
			"owner_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of the owner, must be one of 'organization' or 'application'.",
				Validators: []validator.String{
					stringvalidator.OneOf("organization", "application"),
				},
			},
			"repository_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The SCM provider URL for the repository, only valid for `owner_type` of `application`",
			},
			"base_branch": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The default branch to use.",
			},
			"user_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The user name to use when setting `scm_provider` to `bitbucket`.",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("scm_provider")),
				},
			},
			"remediation_pull_requests_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set to true to enable the Automated Pull Requests feature.",
			},
			"pull_request_commenting_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set to true to enable the Pull Request Commenting feature.",
			},
			"source_control_evaluation_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set to true to enable Nexus IQ triggered source control evaluations.",
			},
			"scm_provider": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The type of SCM Provider, must be one of 'azure, bitbucket, github or gitlab'. This is required when the organization is set to `ROOT_ORGANIZATION_ID`",
				Validators: []validator.String{
					stringvalidator.OneOf("azure", "bitbucket", "github", "gitlab"),
				},
			},
			"token": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The token for use with the SCM Provider 'scm_provider'",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("scm_provider")),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *sourceControlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sourceControlModelResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Create the resource
	sourceControlRequest := r.Client.SourceControlAPI.AddSourceControl(ctx, strings.Trim(plan.OwnerType.String(), "\""), strings.Trim(plan.ID.String(), "\"")).ApiSourceControlDTO(sonatypeiq.ApiSourceControlDTO{
		BaseBranch:                      plan.BaseBranch.ValueStringPointer(),
		SourceControlEvaluationsEnabled: plan.SourceControlEvaluationsEnabled.ValueBoolPointer(),
		PullRequestCommentingEnabled:    plan.PullRequestCommentingEnabled.ValueBoolPointer(),
		RemediationPullRequestsEnabled:  plan.RemediationPullRequestsEnabled.ValueBoolPointer(),
		RepositoryUrl:                   plan.RepositoryUrl.ValueStringPointer(),
		Username:                        plan.UserName.ValueStringPointer(),
		Provider:                        plan.ScmProvider.ValueStringPointer(),
		Token:                           plan.Token.ValueStringPointer(),
	})
	sourceControl, apiResponse, err := sourceControlRequest.Execute()
	if err != nil {
		errorBody, _ := io.ReadAll(apiResponse.Body)
		headers := apiResponse.Request.URL
		resp.Diagnostics.AddError(
			"Error creating Source Control",
			"Could not create Source Control, unexpected error: "+apiResponse.Status+": "+string(errorBody)+headers.String(),
		)
		return
	}

	// Set the state to fully populated data
	plan.ID = types.StringValue(*sourceControl.OwnerId)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Read refreshes the Terraform state with the latest data.
func (r *sourceControlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sourceControlModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)
	// Get refreshed Source Control from IQ
	ownerType := strings.Trim(state.OwnerType.String(), "\"")
	ownerId := strings.Trim(state.ID.String(), "\"")
	tflog.Debug(ctx, fmt.Sprintf("Importing Source Control with owner_type: %s", ownerType))
	tflog.Debug(ctx, fmt.Sprintf("Importing Source Control with owner_id: %s", ownerId))
	sourceControl, apiResponse, err := r.Client.SourceControlAPI.GetSourceControl1(ctx, ownerType, ownerId).Execute()
	if err != nil {
		if apiResponse.StatusCode == 404 {
			tflog.Debug(ctx, "Remove resource from state as no longer exists")
			resp.State.RemoveResource(ctx)
			return
		} else {
			tflog.Error(ctx, "Error Reading IQ Source Control")
			resp.Diagnostics.AddError(
				"Error Reading IQ Source Control",
				"Could not read Source Control for ID "+state.ID.ValueString()+": "+err.Error(),
			)
		}
		return
	} else {
		// Overwrite items with refreshed state
		tflog.Debug(ctx, fmt.Sprintf("Setting imported attributes in state %s", *sourceControl.OwnerId))
		state.ID = types.StringValue(*sourceControl.OwnerId)
		if ownerType == "application" {
			state.RepositoryUrl = types.StringValue(*sourceControl.RepositoryUrl)
		}
		if sourceControl.BaseBranch != nil {
			state.BaseBranch = types.StringValue(*sourceControl.BaseBranch)
		}
		if sourceControl.RemediationPullRequestsEnabled != nil {
			state.RemediationPullRequestsEnabled = types.BoolValue(*sourceControl.RemediationPullRequestsEnabled)
		}
		if sourceControl.PullRequestCommentingEnabled != nil {
			state.PullRequestCommentingEnabled = types.BoolValue(*sourceControl.PullRequestCommentingEnabled)
		}
		if sourceControl.SourceControlEvaluationsEnabled != nil {
			state.SourceControlEvaluationsEnabled = types.BoolValue(*sourceControl.SourceControlEvaluationsEnabled)
		}
		if sourceControl.Provider != nil {
			state.ScmProvider = types.StringValue(*sourceControl.Provider)
		}
		if sourceControl.Token != nil {
			state.Token = types.StringValue(*sourceControl.Token)
		}
		if sourceControl.Username != nil {
			state.UserName = types.StringValue(*sourceControl.Username)
		}
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *sourceControlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sourceControlModelResource
	var state sourceControlModelResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make Update API Call
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)
	sourceControlUpdateRequest := r.Client.SourceControlAPI.UpdateSourceControl(ctx, strings.Trim(state.OwnerType.String(), "\""), strings.Trim(state.ID.String(), "\"")).ApiSourceControlDTO(sonatypeiq.ApiSourceControlDTO{
		BaseBranch:                      plan.BaseBranch.ValueStringPointer(),
		SourceControlEvaluationsEnabled: plan.SourceControlEvaluationsEnabled.ValueBoolPointer(),
		PullRequestCommentingEnabled:    plan.PullRequestCommentingEnabled.ValueBoolPointer(),
		RemediationPullRequestsEnabled:  plan.RemediationPullRequestsEnabled.ValueBoolPointer(),
		RepositoryUrl:                   plan.RepositoryUrl.ValueStringPointer(),
		Username:                        plan.UserName.ValueStringPointer(),
		Provider:                        plan.ScmProvider.ValueStringPointer(),
		Token:                           plan.Token.ValueStringPointer(),
	})
	sourceControl, apiResponse, err := sourceControlUpdateRequest.Execute()
	if err != nil {
		errorBody, _ := io.ReadAll(apiResponse.Body)
		resp.Diagnostics.AddError(
			"Error updating Source Control",
			"Could not update Source Control, unexpected error: "+apiResponse.Status+": "+string(errorBody),
		)
		return
	}
	// Set the state to fully populated data
	plan.ID = types.StringValue(*sourceControl.OwnerId)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sourceControlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sourceControlModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Delete Source Control
	apiResponse, err := r.Client.SourceControlAPI.DeleteSourceControl(ctx, strings.Trim(state.OwnerType.String(), "\""), strings.Trim(state.ID.String(), "\"")).Execute()
	if err != nil {
		errorBody, _ := io.ReadAll(apiResponse.Body)
		resp.Diagnostics.AddError(
			"Error deleting Source Control",
			"Could not delete Source Control, unexpected error: "+apiResponse.Status+": "+string(errorBody),
		)
		return
	}
}

func (r *sourceControlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: owner_type:id. Got: %q", req.ID),
		)
		return
	}
	if idParts[0] != "application" && idParts[0] != "organization" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier prefix",
			fmt.Sprintf("Expected import identifier to start with 'application' or 'organization'. Got: %q", idParts[0]),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Importing Source Control with owner_type: %s and owner_id: %s", idParts[0], idParts[1]))
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner_type"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner_id"), idParts[1])...)
}
