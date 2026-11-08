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
package role_test

import (
	"fmt"
	"terraform-provider-sonatypeiq/internal/provider/common"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Grants the System Administrator role - a global role - to a throwaway group and user, and revokes it again.
func TestAccGlobalRoleMembershipResource(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	groupResource := "sonatypeiq_global_role_membership.group"
	userResource := "sonatypeiq_global_role_membership.user"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfigGlobalRoleMembership(suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(groupResource, "id"),
					resource.TestMatchResourceAttr(groupResource, "role_id", common.ROLE_ID_REGEX),
					resource.TestCheckResourceAttr(groupResource, "group_name", "tf-acc-"+suffix),
					resource.TestCheckResourceAttrSet(groupResource, "last_updated"),
					resource.TestCheckResourceAttr(userResource, "user_name", "tf-acc-"+suffix),
				),
			},
			// No drift
			{
				Config:             testConfigGlobalRoleMembership(suffix),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            groupResource,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			{
				ResourceName:            userResource,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testConfigGlobalRoleMembership(suffix string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
data "sonatypeiq_role" "system_administrator" {
  name = "System Administrator"
}

resource "sonatypeiq_user" "user" {
  username   = "tf-acc-%[1]s"
  password   = "randomthing"
  first_name = "Example"
  last_name  = "User"
  email      = "example@user.tld"
}

resource "sonatypeiq_global_role_membership" "group" {
  role_id    = data.sonatypeiq_role.system_administrator.id
  group_name = "tf-acc-%[1]s"
}

resource "sonatypeiq_global_role_membership" "user" {
  role_id   = data.sonatypeiq_role.system_administrator.id
  user_name = sonatypeiq_user.user.username
}`, suffix)
}
