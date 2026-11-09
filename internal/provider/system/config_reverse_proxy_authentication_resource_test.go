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
	"terraform-provider-sonatypeiq/internal/provider/common"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfigReverseProxyAuthenticationResource(t *testing.T) {
	resourceName := "sonatypeiq_config_reverse_proxy_authentication.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing - with enabled = false (NEVER enable on shared test server!)
			{
				Config: testAccConfigReverseProxyAuthMinimalResource(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_REVERSE_PROXY_AUTHENTICATION_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "username_header"),
					resource.TestCheckNoResourceAttr(resourceName, "csrf_protection_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "logout_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			{
				Config: testAccConfigReverseProxyAuthFullResource(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_REVERSE_PROXY_AUTHENTICATION_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "username_header", "X-Custom-User"),
					resource.TestCheckResourceAttr(resourceName, "csrf_protection_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "logout_url", "https://sso.example.com/logout"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccConfigReverseProxyAuthFullResource(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccConfigReverseProxyAuthMinimalResource() string {
	return utils_test.ProviderConfig + `
resource "sonatypeiq_config_reverse_proxy_authentication" "test" {
  enabled = false
}`
}

func testAccConfigReverseProxyAuthFullResource() string {
	return utils_test.ProviderConfig + `
resource "sonatypeiq_config_reverse_proxy_authentication" "test" {
  enabled                  = false
  username_header          = "X-Custom-User"
  csrf_protection_disabled = true
  logout_url               = "https://sso.example.com/logout"
}`
}
