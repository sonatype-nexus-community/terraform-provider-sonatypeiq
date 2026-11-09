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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfigJiraResource(t *testing.T) {
	randomStr := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "sonatypeiq_config_jira.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccConfigJiraMinimalResource(randomStr),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Jira configuration
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_JIRA_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "url", fmt.Sprintf("https://jira.%s.tld", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "username", randomStr),
					resource.TestCheckResourceAttr(resourceName, "password", "fake-password"),
					resource.TestCheckNoResourceAttr(resourceName, "custom_fields"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			{
				Config: testAccConfigJiraFullResource(randomStr),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Jira configuration with custom fields
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_JIRA_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "url", fmt.Sprintf("https://jira.%s.tld", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "username", randomStr),
					resource.TestCheckResourceAttr(resourceName, "password", "fake-password"),
					resource.TestCheckResourceAttr(resourceName, "custom_fields", `{"field1":{"key1":"value1"},"field2":{"key2":"value2"}}`),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccConfigJiraFullResource(randomStr),
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

func testAccConfigJiraMinimalResource(randomStr string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_config_jira" "test" {
  url      = "https://jira.%s.tld"
  username = "%s"
  password = "fake-password"
}`, randomStr, randomStr)
}

func testAccConfigJiraFullResource(randomStr string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_config_jira" "test" {
  url           = "https://jira.%s.tld"
  username      = "%s"
  password      = "fake-password"
  custom_fields = jsonencode({
    field1 = {
      key1 = "value1"
    }
    field2 = {
      key2 = "value2"
    }
  })
}`, randomStr, randomStr)
}
