/*
 * Copyright (c) 2019-present Sonatype, Inc.
 *)
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

func TestAccConfigProxyServerResource(t *testing.T) {
	randomStr := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "sonatypeiq_config_proxy_server.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccConfigProxyMinimalResource(randomStr),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify mail configuration
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_PROXY_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("smtp.%s.tld", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "port", "8080"),
					resource.TestCheckNoResourceAttr(resourceName, "username"),
					resource.TestCheckNoResourceAttr(resourceName, "password"),
					resource.TestCheckResourceAttr(resourceName, "exclude_hosts.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			{
				Config: testAccConfigProxyFullResource(randomStr),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify mail configuration
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_PROXY_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("smtp.%s.tld", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "port", "465"),
					resource.TestCheckResourceAttr(resourceName, "username", randomStr),
					resource.TestCheckResourceAttr(resourceName, "password", "fake-password"),
					resource.TestCheckResourceAttr(resourceName, "exclude_hosts.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccConfigProxyFullResource(randomStr),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "last_updated"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccConfigProxyMinimalResource(randomStr string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_config_proxy_server" "test" {
  hostname          = "smtp.%s.tld"
  port              = 8080
}`, randomStr)
}

func testAccConfigProxyFullResource(randomStr string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_config_proxy_server" "test" {
  hostname          = "smtp.%s.tld"
  port              = 465
  username 			= "%s"
  password 			= "fake-password"
  exclude_hosts = [
	"*.somewhere.tld",
	"*.elsewhere.tld"
  ]
}`, randomStr, randomStr)
}
