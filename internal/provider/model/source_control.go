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
	"terraform-provider-sonatypeiq/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// SourceControlModelResource
// --------------------------------------------
type SourceControlModelResource struct {
	ID                              types.String `tfsdk:"id"`
	OwnerID                         types.String `tfsdk:"owner_id"`
	OwnerType                       types.String `tfsdk:"owner_type"`
	RepositoryUrl                   types.String `tfsdk:"repository_url"`
	BaseBranch                      types.String `tfsdk:"base_branch"`
	UserName                        types.String `tfsdk:"user_name"`
	RemediationPullRequestsEnabled  types.Bool   `tfsdk:"remediation_pull_requests_enabled"`
	PullRequestCommentingEnabled    types.Bool   `tfsdk:"pull_request_commenting_enabled"`
	SourceControlEvaluationsEnabled types.Bool   `tfsdk:"source_control_evaluation_enabled"`
	Token                           types.String `tfsdk:"token"`
	ScmProvider                     types.String `tfsdk:"scm_provider"`
	LastUpdated                     types.String `tfsdk:"last_updated"`
}

func (m *SourceControlModelResource) MapFromApi(api *sonatypeiq.ApiSourceControlDTO) {
	m.ID = types.StringValue(fmt.Sprintf(common.SCM_CONFIG_ID_FORMAT, m.OwnerType.ValueString(), *api.OwnerId))
	m.OwnerID = types.StringPointerValue(api.OwnerId)
	m.RepositoryUrl = types.StringPointerValue(api.RepositoryUrl)
	m.BaseBranch = types.StringPointerValue(api.BaseBranch)
	m.UserName = types.StringPointerValue(api.Username)
	m.RemediationPullRequestsEnabled = types.BoolPointerValue(api.RemediationPullRequestsEnabled)
	m.PullRequestCommentingEnabled = types.BoolPointerValue(api.PullRequestCommentingEnabled)
	m.SourceControlEvaluationsEnabled = types.BoolPointerValue(api.SourceControlEvaluationsEnabled)
	// Token
	m.ScmProvider = types.StringPointerValue(api.Provider)
}

func (m *SourceControlModelResource) MapToApi() *sonatypeiq.ApiSourceControlDTO {
	api := sonatypeiq.NewApiSourceControlDTOWithDefaults()
	api.OwnerId = m.OwnerID.ValueStringPointer()
	api.RepositoryUrl = m.RepositoryUrl.ValueStringPointer()
	api.BaseBranch = m.BaseBranch.ValueStringPointer()
	api.Username = m.UserName.ValueStringPointer()
	api.RemediationPullRequestsEnabled = m.RemediationPullRequestsEnabled.ValueBoolPointer()
	api.PullRequestCommentingEnabled = m.PullRequestCommentingEnabled.ValueBoolPointer()
	api.SourceControlEvaluationsEnabled = m.SourceControlEvaluationsEnabled.ValueBoolPointer()
	api.Token = m.Token.ValueStringPointer()
	api.Provider = m.ScmProvider.ValueStringPointer()

	// Inject Default Values that only apply at ROOT ORG
	if m.OwnerID.ValueString() == common.ROOT_ORGANIZATION_ID {
		if m.RemediationPullRequestsEnabled.IsNull() {
			api.RemediationPullRequestsEnabled = sonatypeiq.PtrBool(true)
		}
	}

	return api
}
