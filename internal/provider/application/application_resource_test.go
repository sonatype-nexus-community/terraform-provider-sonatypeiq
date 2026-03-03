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
	"terraform-provider-sonatypeiq/internal/provider/common"
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
			// Create
			{
				Config: testAccApplicationResource(appName, "", "data.sonatypeiq_organization.sandbox.id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Application
					resource.TestMatchResourceAttr(resourceName, "id", common.APPLICATION_INTERNAL_ID_REGEX),
					resource.TestCheckResourceAttr(resourceName, "name", appName),
					resource.TestCheckResourceAttr(resourceName, "public_id", appName),
					resource.TestMatchResourceAttr(resourceName, "organization_id", common.ORGANIZATION_ID_REGEX),
					resource.TestCheckResourceAttr(resourceName, "contact_user_name", "admin"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Update and Move Org
			{
				Config: testAccApplicationResource(appName, "2", "sonatypeiq_organization.sub_org.id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "id", common.APPLICATION_INTERNAL_ID_REGEX),
					resource.TestCheckResourceAttr(resourceName, "name", appName+"2"),
					resource.TestCheckResourceAttr(resourceName, "public_id", appName+"2"),
					resource.TestMatchResourceAttr(resourceName, "organization_id", common.ORGANIZATION_ID_REGEX),
					resource.TestCheckResourceAttr(resourceName, "contact_user_name", "admin"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccApplicationResource(appName, "2", "sonatypeiq_organization.sub_org.id"),
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

func testAccApplicationResource(name, update, organization string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

data "sonatypeiq_organization" "root" {
  id = "ROOT_ORGANIZATION_ID"
}

resource "sonatypeiq_organization" "sub_org" {
  name                   = "Sub Organization"
  parent_organization_id = data.sonatypeiq_organization.root.id
}

resource "sonatypeiq_application" "test" {
  name = "%s%s"
  public_id = "%s%s"
  organization_id = %s
  contact_user_name = "admin"
}`, name, update, name, update, organization)
}
