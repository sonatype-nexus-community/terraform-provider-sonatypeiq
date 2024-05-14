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

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationResource(t *testing.T) {

	appName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resourceName := "sonatypeiq_application.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationResource(appName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Application
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", appName),
					resource.TestCheckResourceAttr(resourceName, "public_id", appName),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			{
				Config: testAccApplicationResource(appName, "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", appName+"2"),
					resource.TestCheckResourceAttr(resourceName, "public_id", appName+"2"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccApplicationResource(appName, "2"),
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

func testAccApplicationResource(name string, update string) string {
	return fmt.Sprintf(providerConfig+`
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

resource "sonatypeiq_application" "test" {
  name = "%s%s"
  public_id = "%s%s"
  organization_id = data.sonatypeiq_organization.sandbox.id
}`, name, update, name, update)
}
