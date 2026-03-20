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

func TestAccConfigMailResource(t *testing.T) {
	randomStr := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "sonatypeiq_config_mail.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccConfigMailMinimalResource(randomStr),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify mail configuration
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_MAIL_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("smtp.%s.tld", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "port", "25"),
					resource.TestCheckNoResourceAttr(resourceName, "username"),
					resource.TestCheckNoResourceAttr(resourceName, "password"),
					resource.TestCheckResourceAttr(resourceName, "ssl_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "start_tls_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "system_email", fmt.Sprintf("no-reply@%s.tld", randomStr)),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			{
				Config: testAccConfigMailFullResource(randomStr),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify mail configuration
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_MAIL_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("smtp.%s.tld", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "port", "465"),
					resource.TestCheckResourceAttr(resourceName, "username", randomStr),
					resource.TestCheckResourceAttr(resourceName, "password", "fake-password"),
					resource.TestCheckResourceAttr(resourceName, "ssl_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "start_tls_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "system_email", fmt.Sprintf("no-reply@%s.tld", randomStr)),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccConfigMailFullResource(randomStr),
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
func testAccConfigMailMinimalResource(randomStr string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_config_mail" "test" {
  hostname          = "smtp.%s.tld"
  port              = 25
  ssl_enabled       = false
  start_tls_enabled = false
  system_email      = "no-reply@%s.tld"
}`, randomStr, randomStr)
}

func testAccConfigMailFullResource(randomStr string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_config_mail" "test" {
  hostname          = "smtp.%s.tld"
  port              = 465
  username 			= "%s"
  password 			= "fake-password"
  ssl_enabled       = true
  start_tls_enabled = true
  system_email      = "no-reply@%s.tld"
}`, randomStr, randomStr, randomStr)
}
