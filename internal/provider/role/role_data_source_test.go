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

package role_test

import (
	testutil "terraform-provider-sonatypeiq/internal/provider/testutil"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleDataSourcePre198(t *testing.T) {
	var resourceName = "data.sonatypeiq_role.role_by_name"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: utils_test.ProviderConfig + `data "sonatypeiq_role" "role_by_name" {
					name = "Developer"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "1da70fae1fd54d6cb7999871ebdb9a36"),
					resource.TestCheckResourceAttr(resourceName, "name", "Developer"),
				),
			},
		},
	})
}

func TestAccRoleDataSource(t *testing.T) {
	var resourceName = "data.sonatypeiq_role.role_by_name"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		PreCheck: func() {
			// Not supported prior to NXIQ 198
			testutil.SkipIfNxiqVersionOlderThan(t, 198)
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: utils_test.ProviderConfig + `data "sonatypeiq_role" "role_by_name" {
					name = "Developer"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "1da70fae1fd54d6cb7999871ebdb9a36"),
					resource.TestCheckResourceAttr(resourceName, "name", "Developer"),
					resource.TestCheckResourceAttr(resourceName, "description", "Views all information for their assigned organization or application."),
					resource.TestCheckResourceAttr(resourceName, "built_in", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.admin.access_audit_log", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.admin.view_roles", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.add_applications", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.claim_components", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.edit_access_control", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.edit_iq_elements", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.edit_proprietary_components", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.evaluate_applications", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.evaluate_individual_components", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.manage_automatic_application_creation", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.manage_automatic_scm_configuration", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.view_iq_elements", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.change_licenses", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.change_security_vulnerabilities", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.create_pull_requests", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.review_legal_obligations", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.waive_policy_violations", "false"),
				),
			},
		},
	})
}
