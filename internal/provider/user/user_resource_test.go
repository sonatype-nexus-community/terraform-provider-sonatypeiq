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
	resourceName := "sonatypeiq_user.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserResource(userName, password, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf(common.USER_ID_FORMAT, common.USER_REALM_INTERNAL, userName)),
					resource.TestCheckResourceAttr(resourceName, "username", userName),
					resource.TestCheckResourceAttr(resourceName, "first_name", "Example"),
					resource.TestCheckResourceAttr(resourceName, "last_name", "User"),
					resource.TestCheckResourceAttr(resourceName, "email", fmt.Sprintf("%s@user.tld", userName)),
					resource.TestCheckResourceAttr(resourceName, "realm", common.USER_REALM_INTERNAL),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			{
				Config: testAccUserResource(userName, password, " Esq"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf(common.USER_ID_FORMAT, common.USER_REALM_INTERNAL, userName)),
					resource.TestCheckResourceAttr(resourceName, "username", userName),
					resource.TestCheckResourceAttr(resourceName, "first_name", "Example Esq"),
					resource.TestCheckResourceAttr(resourceName, "last_name", "User"),
					resource.TestCheckResourceAttr(resourceName, "email", fmt.Sprintf("%s@user.tld", userName)),
					resource.TestCheckResourceAttr(resourceName, "realm", common.USER_REALM_INTERNAL),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccUserResource(userName, password, " Esq"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateId:           fmt.Sprintf(common.USER_ID_FORMAT, common.USER_REALM_INTERNAL, userName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "last_updated"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccUserResource(username, password, nameSuffix string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_user" "test" {
  username = "%s"
  password = "%s"
  first_name = "Example%s"
  last_name = "User"
  email = "%s@user.tld"
}`, username, password, nameSuffix, username)
}
