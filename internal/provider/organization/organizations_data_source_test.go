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
	"terraform-provider-sonatypeiq/internal/provider/common"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationsDataSource(t *testing.T) {
	resourceName := "data.sonatypeiq_organizations.orgs"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: utils_test.ProviderConfig + `data "sonatypeiq_organizations" "orgs" {
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "all-organizations"),
					// ROOT ORG
					resource.TestCheckResourceAttr(resourceName, "organizations.0.id", common.ROOT_ORGANIZATION_ID),
					resource.TestCheckResourceAttr(resourceName, "organizations.0.name", "Root Organization"),
					resource.TestCheckNoResourceAttr(resourceName, "organizations.0.parent_organization_id"),
					resource.TestCheckResourceAttr(resourceName, "organizations.0.categories.#", "3"),
					// SANDBOX ORG
					resource.TestMatchResourceAttr(resourceName, "organizations.1.id", common.ORGANIZATION_ID_REGEX),
					resource.TestCheckResourceAttr(resourceName, "organizations.1.name", "Sandbox Organization"),
					resource.TestCheckResourceAttr(resourceName, "organizations.1.parent_organization_id", common.ROOT_ORGANIZATION_ID),
					resource.TestCheckResourceAttr(resourceName, "organizations.1.categories.#", "0"),
				),
			},
		},
	})
}
