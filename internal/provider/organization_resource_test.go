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

func TestAccOrganizationResource(t *testing.T) {

	orgName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOrganizationResource(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Application
					resource.TestCheckResourceAttrSet("sonatypeiq_organization.org", "id"),
					resource.TestCheckResourceAttr("sonatypeiq_organization.org", "name", orgName),
					resource.TestCheckResourceAttr("sonatypeiq_organization.org", "parent_organization_id", "ROOT_ORGANIZATION_ID"),
					resource.TestCheckResourceAttrSet("sonatypeiq_organization.org", "last_updated"),
				),
			},
			// // Update
			// {
			// 	Config: testAccSystemConfigResource(iqUrl, false),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		// Verify Application
			// 		resource.TestCheckResourceAttrSet("sonatypeiq_system_config.config", "id"),
			// 		resource.TestCheckResourceAttr("sonatypeiq_system_config.config", "base_url", iqUrl),
			// 		resource.TestCheckResourceAttr("sonatypeiq_system_config.config", "force_base_url", "false"),
			// 		resource.TestCheckResourceAttrSet("sonatypeiq_system_config.config", "last_updated"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccOrganizationResource(orgName string) string {
	return fmt.Sprintf(providerConfig+`
data "sonatypeiq_organization" "root" {
  id = "ROOT_ORGANIZATION_ID"
}

resource "sonatypeiq_organization" "org" {
  name = "%s"
  parent_organization_id = data.sonatypeiq_organization.root.id
}`, orgName)
}
