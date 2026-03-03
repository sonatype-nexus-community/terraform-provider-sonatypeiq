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

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationDataSource(t *testing.T) {
	var resourceName = "data.sonatypeiq_organization.org"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID
			{
				Config: utils_test.ProviderConfig + fmt.Sprintf(`data "sonatypeiq_organization" "org" {
					id = "%s"
				}`, common.ROOT_ORGANIZATION_ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.ROOT_ORGANIZATION_ID),
					resource.TestCheckResourceAttr(resourceName, "name", "Root Organization"),
					resource.TestCheckNoResourceAttr(resourceName, "parent_organization_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
				),
			},
			// Read by Name
			{
				Config: utils_test.ProviderConfig + `data "sonatypeiq_organization" "org" {
					name = "Sandbox Organization"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "Sandbox Organization"),
					resource.TestCheckResourceAttr(resourceName, "parent_organization_id", common.ROOT_ORGANIZATION_ID),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
				),
			},
		},
	})
}
