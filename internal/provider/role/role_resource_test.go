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
	"fmt"
	"terraform-provider-sonatypeiq/internal/provider/common"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleResource(t *testing.T) {
	randomStr := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "sonatypeiq_role.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRoleResource(randomStr, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "id", common.ROLE_ID_REGEX),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("Test Role %s", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Example Role %s", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "built_in", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.admin.access_audit_log", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.admin.view_roles", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.add_applications", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.claim_components", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.edit_access_control", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.edit_iq_elements", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.edit_proprietary_components", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.evaluate_applications", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.evaluate_individual_components", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.manage_automatic_application_creation", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.manage_automatic_scm_configuration", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.view_iq_elements", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.change_licenses", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.change_security_vulnerabilities", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.create_pull_requests", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.review_legal_obligations", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.waive_policy_violations", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			{
				Config: testAccRoleResource(randomStr, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "id", common.ROLE_ID_REGEX),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("Test Role Complete %s", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Example Role Complete %s", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "built_in", "false"),
					resource.TestCheckResourceAttr(resourceName, "permissions.admin.access_audit_log", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.admin.view_roles", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.add_applications", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.claim_components", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.edit_access_control", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.edit_iq_elements", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.edit_proprietary_components", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.evaluate_applications", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.evaluate_individual_components", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.manage_automatic_application_creation", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.manage_automatic_scm_configuration", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.iq.view_iq_elements", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.change_licenses", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.change_security_vulnerabilities", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.create_pull_requests", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.review_legal_obligations", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.remediation.waive_policy_violations", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccRoleResource(randomStr, false),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccRoleResource(randomStr string, minimal bool) string {
	if minimal {
		return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_role" "test" {
  name = "Test Role %s"
  description = "Example Role %s"
  permissions = {
    admin = {
      access_audit_log = true
      view_roles = false
    }
    iq = {}
    remediation = {}
  }
}`, randomStr, randomStr)
	} else {
		return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_role" "test" {
  name = "Test Role Complete %s"
  description = "Example Role Complete %s"
  permissions = {
    admin = {
      access_audit_log = true
      view_roles = true
    }
    iq = {
	  add_applications = true
	  claim_components = true
	  edit_access_control = true
	  edit_iq_elements = true
	  edit_proprietary_components = true
	  evaluate_applications = true
	  evaluate_individual_components = true
	  manage_automatic_application_creation = true
	  manage_automatic_scm_configuration = true
	  view_iq_elements = true
	}
    remediation = {
	  change_licenses = true
	  change_security_vulnerabilities = true
	  create_pull_requests = true
	  review_legal_obligations = true
	  waive_policy_violations = true
	}
  }
}`, randomStr, randomStr)
	}
}
