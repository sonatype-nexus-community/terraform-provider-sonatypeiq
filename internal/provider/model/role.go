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
	"terraform-provider-sonatypeiq/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// RoleModel
// ------------------------------------------------------------
type RoleModel struct {
	ID          types.String          `tfsdk:"id"`
	Name        types.String          `tfsdk:"name"`
	Description types.String          `tfsdk:"description"`
	BuiltIn     types.Bool            `tfsdk:"built_in"`
	Permissions *RolePermissionsModel `tfsdk:"permissions"`
}

func (m *RoleModel) MapFromApi(api *sonatypeiq.ApiRoleDTO) {
	m.ID = types.StringPointerValue(api.Id)
	m.Name = types.StringPointerValue(api.Name)
	m.Description = types.StringPointerValue(api.Description)
	m.BuiltIn = types.BoolPointerValue(api.BuiltIn)
	if m.Permissions == nil {
		// Will be nil during import
		m.Permissions = &RolePermissionsModel{}
	}
	m.Permissions.MapFromApi(api.PermissionCategories)
}

// RoleModelResource
// ------------------------------------------------------------
type RoleModelResource struct {
	RoleModel
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (m *RoleModelResource) MapFromApi(api *sonatypeiq.ApiRoleDTO) {
	m.RoleModel.MapFromApi(api)
}

func (m *RoleModelResource) MapToApi(includeId bool) *sonatypeiq.ApiRoleDTO {
	api := sonatypeiq.NewApiRoleDTOWithDefaults()
	if includeId {
		api.Id = m.ID.ValueStringPointer()
	}
	api.Name = m.Name.ValueStringPointer()
	api.Description = m.Description.ValueStringPointer()
	// Built-In skipped
	if m.Permissions != nil {
		// Will be nil during import
		api.PermissionCategories = m.Permissions.AsPermissionCategories()
	}
	return api
}

// RolePermissionsModel
// ------------------------------------------------------------
type RolePermissionsModel struct {
	Admin       roleAdminPermissions       `tfsdk:"admin"`
	Iq          roleIqPermissions          `tfsdk:"iq"`
	Remediation roleRemediationPermissions `tfsdk:"remediation"`
}

func (m *RolePermissionsModel) AsPermissionCategories() []sonatypeiq.ApiPermissionCategoryDTO {
	permissionCategories := make([]sonatypeiq.ApiPermissionCategoryDTO, 0)
	permissionCategories = append(permissionCategories, *m.Admin.MapToApi())
	permissionCategories = append(permissionCategories, *m.Iq.MapToApi())
	permissionCategories = append(permissionCategories, *m.Remediation.MapToApi())
	return permissionCategories
}

func (m *RolePermissionsModel) MapFromApi(api []sonatypeiq.ApiPermissionCategoryDTO) {
	for _, pc := range api {
		switch *pc.DisplayName {
		case common.ROLE_PERMISSION_CATEGORY_ADNIN:
			m.Admin.MapFromApi(pc.Permissions)
		case common.ROLE_PERMISSION_CATEGORY_IQ:
			m.Iq.MapFromApi(pc.Permissions)
		case common.ROLE_PERMISSION_CATEGORY_REMEDIATION:
			m.Remediation.MapFromApi(pc.Permissions)
		}
	}
}

// roleAdminPermissions
// ------------------------------------------------------------
type roleAdminPermissions struct {
	AccessAuditLogs types.Bool `tfsdk:"access_audit_log"`
	ViewRoles       types.Bool `tfsdk:"view_roles"`
}

func (m *roleAdminPermissions) MapFromApi(api []sonatypeiq.ApiPermissionDTO) {
	for _, permission := range api {
		switch *permission.Id {
		case common.ROLE_ID_ACCESS_AUDIT_LOGS:
			m.AccessAuditLogs = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_VIEW_ROLES:
			m.ViewRoles = types.BoolPointerValue(permission.Allowed)
		}
	}
}

func (m *roleAdminPermissions) MapToApi() *sonatypeiq.ApiPermissionCategoryDTO {
	return &sonatypeiq.ApiPermissionCategoryDTO{
		DisplayName: sonatypeiq.PtrString(common.ROLE_PERMISSION_CATEGORY_ADNIN),
		Permissions: []sonatypeiq.ApiPermissionDTO{
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_ACCESS_AUDIT_LOGS),
				Allowed: m.AccessAuditLogs.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_VIEW_ROLES),
				Allowed: m.ViewRoles.ValueBoolPointer(),
			},
		},
	}
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

func (m *roleIqPermissions) MapFromApi(api []sonatypeiq.ApiPermissionDTO) {
	for _, permission := range api {
		switch *permission.Id {
		case common.ROLE_ID_ADD_APPLICATIONS:
			m.AddApplications = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_CLAIM_COMPONENTS:
			m.ClaimComponents = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_EDIT_ACCESS_CONTROL:
			m.EditAccessControl = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_EDIT_IQ_ELEMENTS:
			m.EditIqElements = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_EDIT_PROPRIETARY_COMPONENTS:
			m.EditProprietaryComponents = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_EVALUATE_APPLICATIONS:
			m.EvaluateApplications = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_EVALUATE_INDIVIDUAL_COMPONENTS:
			m.EvaluateIndividualComponents = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_MANAGE_AUTOMATIC_APPLICATION_CREATION:
			m.ManageAutomaticApplicationCreation = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_MANAGE_AUTOMATIC_SCM_CONFIGURATION:
			m.ManageAutomaticScmConfiguration = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_VIEW_IQ_ELEMENTS:
			m.ViewIqElements = types.BoolPointerValue(permission.Allowed)
		}
	}
}

func (m *roleIqPermissions) MapToApi() *sonatypeiq.ApiPermissionCategoryDTO {
	return &sonatypeiq.ApiPermissionCategoryDTO{
		DisplayName: sonatypeiq.PtrString(common.ROLE_PERMISSION_CATEGORY_IQ),
		Permissions: []sonatypeiq.ApiPermissionDTO{
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_ADD_APPLICATIONS),
				Allowed: m.AddApplications.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_CLAIM_COMPONENTS),
				Allowed: m.ClaimComponents.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_EDIT_ACCESS_CONTROL),
				Allowed: m.EditAccessControl.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_EDIT_IQ_ELEMENTS),
				Allowed: m.EditIqElements.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_EDIT_PROPRIETARY_COMPONENTS),
				Allowed: m.EditProprietaryComponents.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_EVALUATE_APPLICATIONS),
				Allowed: m.EvaluateApplications.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_EVALUATE_INDIVIDUAL_COMPONENTS),
				Allowed: m.EvaluateIndividualComponents.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_MANAGE_AUTOMATIC_APPLICATION_CREATION),
				Allowed: m.ManageAutomaticApplicationCreation.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_MANAGE_AUTOMATIC_SCM_CONFIGURATION),
				Allowed: m.ManageAutomaticScmConfiguration.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_VIEW_IQ_ELEMENTS),
				Allowed: m.ViewIqElements.ValueBoolPointer(),
			},
		},
	}
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

func (m *roleRemediationPermissions) MapFromApi(api []sonatypeiq.ApiPermissionDTO) {
	for _, permission := range api {
		switch *permission.Id {
		case common.ROLE_ID_CHANGE_LICENSES:
			m.ChangeLicenses = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_CHANGE_SECURITY_VULNERABILITIES:
			m.ChangeSecurityVulnerabilities = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_CREATE_PULL_REQUESTS:
			m.CreatePullRequests = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_REVIEW_LEGAL_OBLIGATIONS:
			m.ReviewLegalObligations = types.BoolPointerValue(permission.Allowed)
		case common.ROLE_ID_WAIVE_POLICY_VIOLATIONS:
			m.WaivePolicyViolations = types.BoolPointerValue(permission.Allowed)
		}
	}
}

func (m *roleRemediationPermissions) MapToApi() *sonatypeiq.ApiPermissionCategoryDTO {
	return &sonatypeiq.ApiPermissionCategoryDTO{
		DisplayName: sonatypeiq.PtrString(common.ROLE_PERMISSION_CATEGORY_REMEDIATION),
		Permissions: []sonatypeiq.ApiPermissionDTO{
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_CHANGE_LICENSES),
				Allowed: m.ChangeLicenses.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_CHANGE_SECURITY_VULNERABILITIES),
				Allowed: m.ChangeSecurityVulnerabilities.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_CREATE_PULL_REQUESTS),
				Allowed: m.CreatePullRequests.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_REVIEW_LEGAL_OBLIGATIONS),
				Allowed: m.ReviewLegalObligations.ValueBoolPointer(),
			},
			{
				Id:      sonatypeiq.PtrString(common.ROLE_ID_WAIVE_POLICY_VIOLATIONS),
				Allowed: m.WaivePolicyViolations.ValueBoolPointer(),
			},
		},
	}
}
