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
	"terraform-provider-sonatypeiq/internal/provider/model"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccOrganizationCategoryResource(t *testing.T) {
	randomId := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "sonatypeiq_application_category.cat"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationCategoryResource(randomId, "", model.ColorDarkBlue.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Application
					resource.TestMatchResourceAttr(resourceName, "id", common.ORGANIZATION_ID_REGEX),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("app-cat-%s", randomId)),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("desc-%s", randomId)),
					resource.TestCheckResourceAttr(resourceName, "organization_id", common.ROOT_ORGANIZATION_ID),
					resource.TestCheckResourceAttr(resourceName, "color", model.ColorDarkBlue.String()),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Update
			{
				Config: testAccApplicationCategoryResource(randomId, "2", model.ColorDarkGreen.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Application
					resource.TestMatchResourceAttr(resourceName, "id", common.ORGANIZATION_ID_REGEX),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("app-cat-%s2", randomId)),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("desc-%s2", randomId)),
					resource.TestCheckResourceAttr(resourceName, "organization_id", common.ROOT_ORGANIZATION_ID),
					resource.TestCheckResourceAttr(resourceName, "color", model.ColorDarkGreen.String()),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccApplicationCategoryResource(randomId, "2", model.ColorDarkGreen.String()),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					primary := s.RootModule().Resources[resourceName].Primary
					id := primary.ID
					organizationId := primary.Attributes["organization_id"]
					return fmt.Sprintf("%s,%s", organizationId, id), nil
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccApplicationCategoryResource(randomId, seq, color string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_application_category" "cat" {
  name = "app-cat-%s%s"
  description = "desc-%s%s"
  organization_id = "%s"
  color = "%s"
}`, randomId, seq, randomId, seq, common.ROOT_ORGANIZATION_ID, color)
}
