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

package organization_test

import (
	"fmt"
	"terraform-provider-sonatypeiq/internal/provider/common"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationResource(t *testing.T) {

	orgName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "sonatypeiq_organization.org"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOrganizationResource(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Application
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", orgName),
					resource.TestCheckResourceAttr(resourceName, "parent_organization_id", common.ROOT_ORGANIZATION_ID),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccOrganizationResource(orgName string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
data "sonatypeiq_organization" "root" {
  id = "`+common.ROOT_ORGANIZATION_ID+`"
}

resource "sonatypeiq_organization" "org" {
  name = "%s"
  parent_organization_id = data.sonatypeiq_organization.root.id
}`, orgName)
}
