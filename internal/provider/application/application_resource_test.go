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

package application_test

import (
	"fmt"
	"regexp"
	"strings"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationResource(t *testing.T) {

	appName := `TFACC` + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resourceName := "sonatypeiq_application.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationResource(appName, "", "data.sonatypeiq_organization.sandbox.id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Application
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", appName),
					resource.TestCheckResourceAttr(resourceName, "public_id", appName),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			{
				Config: testAccApplicationResource(appName, "2", "data.sonatypeiq_organization.sandbox.id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", appName+"2"),
					resource.TestCheckResourceAttr(resourceName, "public_id", appName+"2"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccApplicationResource(appName, "2", "data.sonatypeiq_organization.sandbox.id"),
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

func TestAccApplicationResourceMoveOrganization(t *testing.T) {
	appName := `TFACC` + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomOrgId := strings.ToLower(acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum))
	organizationIdRegex, _ := regexp.Compile(`^[a-z0-9]{32}$`)
	resourceName := "sonatypeiq_application.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationWithSubOrgResource(appName, "", "data.sonatypeiq_organization.sandbox.id", randomOrgId),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Application created in Test Sub Org
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
				),
			},
			{
				Config:      testAccApplicationWithSubOrgResource(appName, "", fmt.Sprintf("\"non-existent-%s\"", randomOrgId), randomOrgId),
				ExpectError: regexp.MustCompile("Could not move the application"),
			},
			{
				Config: testAccApplicationWithSubOrgResource(appName, "", "sonatypeiq_organization.test_sub_org.id", randomOrgId),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Application moved to sandbox organization
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestMatchResourceAttr(resourceName, "organization_id", organizationIdRegex),
				),
			},
		},
	})
}

func testAccApplicationResource(name string, update string, organization string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

data "sonatypeiq_organization" "root" {
  id = "ROOT_ORGANIZATION_ID"
}

resource "sonatypeiq_application" "test" {
  name = "%s%s"
  public_id = "%s%s"
  organization_id = %s
}`, name, update, name, update, organization)
}

func testAccApplicationWithSubOrgResource(name string, update string, organization string, randomId string) string {
	return fmt.Sprintf(testAccApplicationResource(name, update, organization)+`
resource "sonatypeiq_organization" "test_sub_org" {
  name = "test-sub-org-%s"
  parent_organization_id = data.sonatypeiq_organization.root.id
}`, randomId)
}
