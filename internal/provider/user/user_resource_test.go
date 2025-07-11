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

package user_test

import (
	"fmt"
	"terraform-provider-sonatypeiq/internal/provider/common"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {

	userName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	password := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserResource(userName, password, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Application
					resource.TestCheckResourceAttrSet("sonatypeiq_user.user1", "id"),
					resource.TestCheckResourceAttr("sonatypeiq_user.user1", "username", userName),
					resource.TestCheckResourceAttr("sonatypeiq_user.user1", "first_name", "Example"),
					resource.TestCheckResourceAttr("sonatypeiq_user.user1", "last_name", "User"),
					resource.TestCheckResourceAttr("sonatypeiq_user.user1", "email", fmt.Sprintf("%s@user.tld", userName)),
					resource.TestCheckResourceAttr("sonatypeiq_user.user1", "realm", common.DEFAULT_USER_REALM),
					resource.TestCheckResourceAttrSet("sonatypeiq_user.user1", "last_updated"),
				),
			},
			{
				Config: testAccUserResource(userName, password, " Esq"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sonatypeiq_user.user1", "id"),
					resource.TestCheckResourceAttr("sonatypeiq_user.user1", "username", userName),
					resource.TestCheckResourceAttr("sonatypeiq_user.user1", "first_name", "Example Esq"),
					resource.TestCheckResourceAttr("sonatypeiq_user.user1", "last_name", "User"),
					resource.TestCheckResourceAttr("sonatypeiq_user.user1", "email", fmt.Sprintf("%s@user.tld", userName)),
					resource.TestCheckResourceAttr("sonatypeiq_user.user1", "realm", common.DEFAULT_USER_REALM),
					resource.TestCheckResourceAttrSet("sonatypeiq_user.user1", "last_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccUserResource(username string, password string, nameSuffix string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_user" "user1" {
  username = "%s"
  password = "%s"
  first_name = "Example%s"
  last_name = "User"
  email = "%s@user.tld"
}`, username, password, nameSuffix, username)
}
