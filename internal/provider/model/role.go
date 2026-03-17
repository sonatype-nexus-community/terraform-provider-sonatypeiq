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

// RoleModel
// ------------------------------------------------------------
type RoleModel struct {
	ID          types.String         `tfsdk:"id"`
	Name        types.String         `tfsdk:"name"`
	Description types.String         `tfsdk:"description"`
	BuiltIn     types.Bool           `tfsdk:"built_in"`
	Permissions RolePermissionsModel `tfsdk:"permissions"`
}

func (m *RoleModel) MapFromApi(api *sonatypeiq.ApiRoleDTO) {
	m.ID = types.StringPointerValue(api.Id)
	m.Name = types.StringPointerValue(api.Name)
}

// RoleModelResource
// ------------------------------------------------------------
type RoleModelResource struct {
	RoleModel
	LastUpdated types.String `tfsdk:"last_updated"`
}

// RolePermissionsModel
// ------------------------------------------------------------
type RolePermissionsModel struct {
	Admin       roleAdminPermissions       `tfsdk:"admin"`
	Iq          roleIqPermissions          `tfsdk:"admin"`
	Remediation roleRemediationPermissions `tfsdk:"admin"`
}

// roleAdminPermissions
// ------------------------------------------------------------
type roleAdminPermissions struct {
	AccessAuditLogs types.Bool `tfsdk:"access_audit_log"`
	ViewRoles       types.Bool `tfsdk:"view_roles"`
}

// roleIqPermissions
// ------------------------------------------------------------
type roleIqPermissions struct {
	AddApplications                    types.Bool `tfsdk:"add_applications"`
	ClaimComponents                    types.Bool `tfsdk:"claim_components"`
	EditAccessControl                  types.Bool `tfsdk:"edit_access_control"`
	EditIqElements                     types.Bool `tfsdk:"edit_iq_elements"`
	EditProprietaryComponents          types.Bool `tfsdk:"edit_proprietary_components"`
	EvaluateApplications               types.Bool `tfsdk:"evaluate_applications"`
	EvaluateIndividualComponents       types.Bool `tfsdk:"evaluate_individual_components"`
	ManageAutomaticApplicationCreation types.Bool `tfsdk:"manage_automatic_application_creation"`
	ManageAutomaticScmConfiguration    types.Bool `tfsdk:"manage_automatic_scm_configuration"`
	ViewIqElements                     types.Bool `tfsdk:"view_iq_elements"`
}

// roleRemediationPermissions
// ------------------------------------------------------------
type roleRemediationPermissions struct {
	ChangeLicenses                types.Bool `tfsdk:"change_licenses"`
	ChangeSecurityVulnerabilities types.Bool `tfsdk:"change_security_vulnerabilities"`
	CreatePullRequests            types.Bool `tfsdk:"create_pull_requests"`
	ReviewLegalObligations        types.Bool `tfsdk:"review_legal_obligations"`
	WaivePolicyViolations         types.Bool `tfsdk:"waive_policy_violations"`
}
