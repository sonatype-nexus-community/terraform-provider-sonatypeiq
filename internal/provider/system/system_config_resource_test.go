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
	"os"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSystemConfigResource(t *testing.T) {

	// appName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	iqUrl := os.Getenv("IQ_SERVER_URL") + "/"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSystemConfigResource(iqUrl, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Application
					resource.TestCheckResourceAttrSet("sonatypeiq_system_config.config", "id"),
					resource.TestCheckResourceAttr("sonatypeiq_system_config.config", "base_url", iqUrl),
					resource.TestCheckResourceAttr("sonatypeiq_system_config.config", "force_base_url", "true"),
					resource.TestCheckResourceAttrSet("sonatypeiq_system_config.config", "last_updated"),
				),
			},
			// // Update
			// Can't test this in parallel against a Single IQ Server
			// {
			// 	Config: testAccSystemConfigResource(iqUrl, false),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		// Verify Application
			// 		resource.TestCheckResourceAttrSet("sonatypeiq_system_config.config", "id"),
			// 		resource.TestCheckResourceAttr("sonatypeiq_system_config.config", "base_url", iqUrl),
			// 		resource.TestCheckResourceAttr("sonatypeiq_system_config.config", "force_base_url", "false"),
			// 		resource.TestCheckResourceAttrSet("sonatypeiq_system_config.config", "last_updated"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSystemConfigResource(url string, force bool) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_system_config" "config" {
  base_url = "%s"
  force_base_url = %t
}`, url, force)
}
