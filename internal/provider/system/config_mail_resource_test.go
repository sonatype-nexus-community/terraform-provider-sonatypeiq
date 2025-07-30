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
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfigMailResource(t *testing.T) {
	resourceName := "sonatypeiq_config_mail.mail_config"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccConfigMailResource(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify mail configuration
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "port", "25"),
					resource.TestCheckResourceAttr(resourceName, "ssl_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "start_tls_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "system_email", "no-reply@my-domain.tld"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccConfigMailResource() string {
	return fmt.Sprintf(utils_test.ProviderConfig + `
resource "sonatypeiq_config_mail" "mail_config" {
  hostname          = "smtp.my-domain.tld"
  port              = 25
  ssl_enabled       = false
  start_tls_enabled = false
  system_email      = "no-reply@my-domain.tld"
}`)
}
