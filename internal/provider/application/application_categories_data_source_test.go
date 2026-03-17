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
	"testing"

	"terraform-provider-sonatypeiq/internal/provider/common"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationCategoriesDataSource(t *testing.T) {
	resourceName := "data.sonatypeiq_application_categories.categories"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: utils_test.ProviderConfig + `data "sonatypeiq_application_categories" "categories" {
					organization_id = "` + common.ROOT_ORGANIZATION_ID + `"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("application-categories-%s", common.ROOT_ORGANIZATION_ID)),
					resource.TestCheckResourceAttr(resourceName, "organization_id", common.ROOT_ORGANIZATION_ID),
					resource.TestCheckResourceAttr(resourceName, "categories.#", "3"),
					// Distributed
					resource.TestCheckResourceAttrSet(resourceName, "categories.0.id"),
					resource.TestCheckResourceAttr(resourceName, "categories.0.name", "Distributed"),
					resource.TestCheckResourceAttr(resourceName, "categories.0.description", "Applications that are provided for consumption outside the company"),
					resource.TestCheckResourceAttr(resourceName, "categories.0.organization_id", common.ROOT_ORGANIZATION_ID),
					resource.TestCheckResourceAttr(resourceName, "categories.0.color", "yellow"),
					// Hosted
					resource.TestCheckResourceAttrSet(resourceName, "categories.1.id"),
					resource.TestCheckResourceAttr(resourceName, "categories.1.name", "Hosted"),
					resource.TestCheckResourceAttr(resourceName, "categories.1.description", "Applications that are hosted such as services or software as a service."),
					resource.TestCheckResourceAttr(resourceName, "categories.1.organization_id", common.ROOT_ORGANIZATION_ID),
					resource.TestCheckResourceAttr(resourceName, "categories.1.color", "light-purple"),
					// Internal
					resource.TestCheckResourceAttrSet(resourceName, "categories.2.id"),
					resource.TestCheckResourceAttr(resourceName, "categories.2.name", "Internal"),
					resource.TestCheckResourceAttr(resourceName, "categories.2.description", "Applications that are used only by your employees"),
					resource.TestCheckResourceAttr(resourceName, "categories.2.organization_id", common.ROOT_ORGANIZATION_ID),
					resource.TestCheckResourceAttr(resourceName, "categories.2.color", "dark-green"),
				),
			},
		},
	})
}
