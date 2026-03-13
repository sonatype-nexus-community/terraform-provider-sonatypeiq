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

package system_test

import (
	"fmt"
	"terraform-provider-sonatypeiq/internal/provider/common"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfigCrowdResource(t *testing.T) {
	randomStr := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "sonatypeiq_config_crowd.crowd"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccConfigCrowdResource(randomStr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_CROWD_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "server_url", fmt.Sprintf("http://something2/%s", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "application_name", fmt.Sprintf("name-%s", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "application_password", "something"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccConfigCrowdResource(randomStr),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"application_password", "last_updated"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccConfigCrowdResource(randomStr string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_config_crowd" "crowd" {
  server_url = "http://something2/%s"
  application_name = "name-%s"
  application_password = "something"
}`, randomStr, randomStr)
}
