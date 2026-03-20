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

package role

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	"github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &roleDataSource{}
	_ datasource.DataSourceWithConfigure = &roleDataSource{}
)

// RoleDataSource is a helper function to simplify the provider implementation.
func RoleDataSource() datasource.DataSource {
	return &roleDataSource{}
}

// roleDataSource is the data source implementation.
type roleDataSource struct {
	common.BaseDataSource
}

// Metadata returns the data source type name.
func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Schema defines the schema for the data source.
func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Description: "Use this data source to get a role",
		Attributes: map[string]tfschema.Attribute{
			"id":          schema.DataSourceComputedString("Internal ID of this Role"),
			"name":        schema.DataSourceRequiredString("The role name"),
			"description": schema.DataSourceComputedString("Role Description"),
			"built_in":    schema.DataSourceComputedBool("Whether this is a built-in Role in Sonatype IQ"),
			"permissions": schema.DataSourceComputedSingleNestedAttribute(
				"Permissions for this Role",
				map[string]tfschema.Attribute{
					"admin": schema.DataSourceComputedSingleNestedAttribute(
						"Administrator Permmissions",
						map[string]tfschema.Attribute{
							"access_audit_log": schema.DataSourceComputedBool("Access to Audit Logs"),
							"view_roles":       schema.DataSourceComputedBool("View all Roles"),
						},
					),
					"iq": schema.DataSourceComputedSingleNestedAttribute(
						"Sonatype IQ Permmissions",
						map[string]tfschema.Attribute{
							"add_applications":                      schema.DataSourceComputedBool("Can add Applications"),
							"claim_components":                      schema.DataSourceComputedBool("Can Claim Components"),
							"edit_access_control":                   schema.DataSourceComputedBool("Can edit Access Control"),
							"edit_iq_elements":                      schema.DataSourceComputedBool("Can edit IQ Elements"),
							"edit_proprietary_components":           schema.DataSourceComputedBool("Can edit Proprietary Components"),
							"evaluate_applications":                 schema.DataSourceComputedBool("Can Evaluate Applications"),
							"evaluate_individual_components":        schema.DataSourceComputedBool("Can Evaluate Individual Components"),
							"manage_automatic_application_creation": schema.DataSourceComputedBool("Can manage Automatic Application creation"),
							"manage_automatic_scm_configuration":    schema.DataSourceComputedBool("Can manage Automatic SCM Configuration"),
							"view_iq_elements":                      schema.DataSourceComputedBool("Can view IQ Elements"),
						},
					),
					"remediation": schema.DataSourceComputedSingleNestedAttribute(
						"Remediation Permmissions",
						map[string]tfschema.Attribute{
							"change_licenses":                 schema.DataSourceComputedBool("Can change Licenses"),
							"change_security_vulnerabilities": schema.DataSourceComputedBool("Can change Security Vulnerabilities"),
							"create_pull_requests":            schema.DataSourceComputedBool("Can create Pull Requests"),
							"review_legal_obligations":        schema.DataSourceComputedBool("Can review Legal Obligations"),
							"waive_policy_violations":         schema.DataSourceComputedBool("Can Waive Policy Violations"),
						},
					),
				},
			),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.RoleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := d.Client.RolesAPI.GetRoles(d.AuthContext(ctx)).Execute()

	if err != nil {
		errors.HandleAPIError(
			common.ERR_FAILED_READING_ROLES,
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.AddAPIErrorDiagnostic(&resp.Diagnostics, "read", "Roles", httpResponse, err)
		return
	}

	for _, apiRole := range apiResponse.Roles {
		if *apiRole.Name == data.Name.ValueString() {
			// Match
			apiRoleResponse, httpResponse, err := d.Client.RolesAPI.GetRoleById(d.AuthContext(ctx), *apiRole.Id).Execute()

			if err != nil {
				errors.HandleAPIError(
					common.ERR_FAILED_READING_ROLE_BY_ID,
					&err,
					httpResponse,
					&resp.Diagnostics,
				)
				return
			} else if httpResponse.StatusCode != http.StatusOK {
				errors.AddAPIErrorDiagnostic(&resp.Diagnostics, "read", "Role", httpResponse, err)
				return
			}
			data.MapFromApi(apiRoleResponse)
			break
		}
	}

	// No Role found
	if data.ID.IsNull() {
		errors.AddValidationDiagnostic(&resp.Diagnostics, "Role", fmt.Sprintf("Role '%s' does not exist", data.Name.ValueString()))
		return
	}

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
