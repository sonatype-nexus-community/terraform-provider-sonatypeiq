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
	// TODO: Missing LastUpdated
}

func (m *SourceControlModelResource) MapFromApi(api *sonatypeiq.ApiSourceControlDTO) {
	m.ID = types.StringPointerValue(api.OwnerId)
	// if m.OwnerType.ValueString() == common.OWNER_TYPE_APPLICATION {
	m.RepositoryUrl = types.StringPointerValue(api.RepositoryUrl)
	// }
	m.BaseBranch = types.StringPointerValue(api.BaseBranch)
	m.UserName = types.StringPointerValue(api.Username)
	m.RemediationPullRequestsEnabled = types.BoolPointerValue(api.RemediationPullRequestsEnabled)
	m.PullRequestCommentingEnabled = types.BoolPointerValue(api.PullRequestCommentingEnabled)
	m.SourceControlEvaluationsEnabled = types.BoolPointerValue(api.SourceControlEvaluationsEnabled)
	// Token
	m.ScmProvider = types.StringPointerValue(api.Provider)
}

func (m *SourceControlModelResource) MapToApi(api *sonatypeiq.ApiSourceControlDTO) {
	api.OwnerId = m.ID.ValueStringPointer()
	api.RepositoryUrl = m.RepositoryUrl.ValueStringPointer()
	api.BaseBranch = m.BaseBranch.ValueStringPointer()
	api.Username = m.UserName.ValueStringPointer()
	api.RemediationPullRequestsEnabled = m.RemediationPullRequestsEnabled.ValueBoolPointer()
	api.PullRequestCommentingEnabled = m.PullRequestCommentingEnabled.ValueBoolPointer()
	api.SourceControlEvaluationsEnabled = m.SourceControlEvaluationsEnabled.ValueBoolPointer()
	api.Token = m.Token.ValueStringPointer()
	api.Provider = m.ScmProvider.ValueStringPointer()
}
