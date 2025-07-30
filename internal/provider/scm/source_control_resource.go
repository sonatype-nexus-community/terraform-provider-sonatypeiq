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

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// var _ resource.ResourceWithImportState = &sourceControlResource{}

type sourceControlResource struct {
	common.BaseResourceWithImport
}

// type sourceControlModelResource struct {
// 	ID                              types.String `tfsdk:"owner_id"`
// 	OwnerType                       types.String `tfsdk:"owner_type"`
// 	RepositoryUrl                   types.String `tfsdk:"repository_url"`
// 	BaseBranch                      types.String `tfsdk:"base_branch"`
// 	UserName                        types.String `tfsdk:"user_name"`
// 	RemediationPullRequestsEnabled  types.Bool   `tfsdk:"remediation_pull_requests_enabled"`
// 	PullRequestCommentingEnabled    types.Bool   `tfsdk:"pull_request_commenting_enabled"`
// 	SourceControlEvaluationsEnabled types.Bool   `tfsdk:"source_control_evaluation_enabled"`
// 	Token                           types.String `tfsdk:"token"`
// 	ScmProvider                     types.String `tfsdk:"scm_provider"` // This is provider in the rest API but provider is a reserved keyword
// }

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
				Computed:            true,
				MarkdownDescription: "Set to true to enable the Automated Pull Requests feature.",
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *sourceControlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve Plan
	var plan model.SourceControlModelResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Call to Create API
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	apiDto := sonatypeiq.NewApiSourceControlDTOWithDefaults()
	plan.MapToApi(apiDto)
	_, httpResponse, err := r.Client.SourceControlAPI.AddSourceControl(
		ctx,
		strings.Trim(plan.OwnerType.ValueString(), "\""),
		strings.Trim(plan.ID.ValueString(), "\""),
	).ApiSourceControlDTO(*apiDto).Execute()

	// Handle Error
	if err != nil || httpResponse.StatusCode != http.StatusOK {
		common.HandleApiError(
			"Error creating Source Control configuration",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// Set the state to fully populated data
	// plan.ID = types.StringValue(*sourceControl.OwnerId)
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
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Call API to Read
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Read API Call
	apiResponse, httpResponse, err := r.Client.SourceControlAPI.GetSourceControl1(
		ctx,
		strings.Trim(state.OwnerType.String(), "\""),
		strings.Trim(state.ID.String(), "\""),
	).Execute()

	if err != nil {
		if httpResponse.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			common.HandleApiWarning(
				"Source Control configuration does not exist",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		} else {
			common.HandleApiError(
				"Error Reading Source Control configuration",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		}
		return
	}

	// Update State
	state.MapFromApi(apiResponse)
	// state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *sourceControlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan & state
	var plan model.SourceControlModelResource
	var state model.SourceControlModelResource

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting plan data has errors: %v", resp.Diagnostics.Errors()))
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting state data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Make Update API Call
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Call API to Update
	apiDto := sonatypeiq.NewApiSourceControlDTOWithDefaults()
	plan.MapToApi(apiDto)
	_, httpResponse, err := r.Client.SourceControlAPI.UpdateSourceControl(
		ctx,
		strings.Trim(state.OwnerType.String(), "\""),
		strings.Trim(state.ID.String(), "\""),
	).ApiSourceControlDTO(*apiDto).Execute()

	// Handle Error
	if err != nil || httpResponse.StatusCode != http.StatusOK {
		common.HandleApiError(
			"Error updating Source Control configuration",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sourceControlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state model.SourceControlModelResource

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting state data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Call API to Update
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Delete Source Control
	httpResponse, err := r.Client.SourceControlAPI.DeleteSourceControl(
		ctx,
		strings.Trim(state.OwnerType.String(), "\""),
		strings.Trim(state.ID.String(), "\""),
	).Execute()

	// Handle Error
	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		common.HandleApiError(
			"Error removing Source Control configuration",
			&err,
			httpResponse,
			&resp.Diagnostics,
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
