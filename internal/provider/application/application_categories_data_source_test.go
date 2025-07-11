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
	"testing"

	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationCategoriesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: utils_test.ProviderConfig + `data "sonatypeiq_application_categories" "root_cats" {
					organization_id = "ROOT_ORGANIZATION_ID"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sonatypeiq_application_categories.root_cats", "id"),
					resource.TestCheckResourceAttr("data.sonatypeiq_application_categories.root_cats", "organization_id", "ROOT_ORGANIZATION_ID"),
					resource.TestCheckResourceAttrSet("data.sonatypeiq_application_categories.root_cats", "categories.#"),
				),
			},
		},
	})
}
