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
	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	sharedutil "github.com/sonatype-nexus-community/terraform-provider-shared/util"
)

type SourceControlModelResource struct {
	ID                              types.String `tfsdk:"owner_id"` // TODO: Should be OwnerID
	OwnerType                       types.String `tfsdk:"owner_type"`
	RepositoryUrl                   types.String `tfsdk:"repository_url"`
	BaseBranch                      types.String `tfsdk:"base_branch"`
	UserName                        types.String `tfsdk:"user_name"`
	RemediationPullRequestsEnabled  types.Bool   `tfsdk:"remediation_pull_requests_enabled"`
	PullRequestCommentingEnabled    types.Bool   `tfsdk:"pull_request_commenting_enabled"`
	SourceControlEvaluationsEnabled types.Bool   `tfsdk:"source_control_evaluation_enabled"`
	Token                           types.String `tfsdk:"token"`
	ScmProvider                     types.String `tfsdk:"scm_provider"` // This is provider in the rest API but provider is a reserved keyword
	LastUpdated                     types.String `tfsdk:"last_updated"`
}

func (m *SourceControlModelResource) MapFromApi(api *sonatypeiq.ApiSourceControlDTO) {
	m.ID = sharedutil.StringPtrToValue(api.OwnerId)
	// if m.OwnerType.ValueString() == common.OWNER_TYPE_APPLICATION {
	m.RepositoryUrl = sharedutil.StringPtrToValue(api.RepositoryUrl)
	// }
	m.BaseBranch = sharedutil.StringPtrToValue(api.BaseBranch)
	m.UserName = sharedutil.StringPtrToValue(api.Username)
	m.RemediationPullRequestsEnabled = sharedutil.BoolPtrToValue(api.RemediationPullRequestsEnabled)
	m.PullRequestCommentingEnabled = sharedutil.BoolPtrToValue(api.PullRequestCommentingEnabled)
	m.SourceControlEvaluationsEnabled = sharedutil.BoolPtrToValue(api.SourceControlEvaluationsEnabled)
	// Token
	m.ScmProvider = sharedutil.StringPtrToValue(api.Provider)
}

func (m *SourceControlModelResource) MapToApi(api *sonatypeiq.ApiSourceControlDTO) {
	api.OwnerId = sharedutil.StringToPtr(m.ID.ValueString())
	api.RepositoryUrl = sharedutil.StringToPtr(m.RepositoryUrl.ValueString())
	api.BaseBranch = sharedutil.StringToPtr(m.BaseBranch.ValueString())
	api.Username = sharedutil.StringToPtr(m.UserName.ValueString())
	api.RemediationPullRequestsEnabled = sharedutil.BoolToPtr(m.RemediationPullRequestsEnabled.ValueBool())
	api.PullRequestCommentingEnabled = sharedutil.BoolToPtr(m.PullRequestCommentingEnabled.ValueBool())
	api.SourceControlEvaluationsEnabled = sharedutil.BoolToPtr(m.SourceControlEvaluationsEnabled.ValueBool())
	api.Token = sharedutil.StringToPtr(m.Token.ValueString())
	api.Provider = sharedutil.StringToPtr(m.ScmProvider.ValueString())
}
