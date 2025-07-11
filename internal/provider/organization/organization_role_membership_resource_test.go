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

package organization_test

import (
	"fmt"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationRoleMembershipResource(t *testing.T) {

	userName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(utils_test.ProviderConfig+`
		 data "sonatypeiq_organization" "sandbox" {
		   name = "Sandbox Organization"
		 }
 
		 data "sonatypeiq_role" "developer" {
		   name = "Developer"
		 }
 
		 resource "sonatypeiq_user" "user" {
		   username   = "%s"
		   password   = "randomthing"
		   first_name = "Example"
		   last_name  = "User"
		   email      = "example@user.tld"
		 }
 
		 resource "sonatypeiq_organization_role_membership" "test" {
		   role_id         = data.sonatypeiq_role.developer.id
		   organization_id = data.sonatypeiq_organization.sandbox.id
		   user_name       = sonatypeiq_user.user.username
		 }
 
		 `, userName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify application role membership
					resource.TestCheckResourceAttrSet("sonatypeiq_organization_role_membership.test", "id"),
					resource.TestCheckResourceAttr("sonatypeiq_organization_role_membership.test", "user_name", userName),
				),
			},
			{
				Config: fmt.Sprintf(utils_test.ProviderConfig+`
		 data "sonatypeiq_organization" "sandbox" {
		   name = "Sandbox Organization"
		 }
 
		 data "sonatypeiq_role" "owner" {
		   name = "Owner"
		 }
 
		 resource "sonatypeiq_user" "user" {
		   username   = "%s"
		   password   = "randomthing"
		   first_name = "Example"
		   last_name  = "User"
		   email      = "example@user.tld"
		 }
 
		 resource "sonatypeiq_organization_role_membership" "test" {
		   role_id         = data.sonatypeiq_role.owner.id
		   organization_id = data.sonatypeiq_organization.sandbox.id
		   user_name       = sonatypeiq_user.user.username
		 }
 
		 `, userName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify application role membership
					resource.TestCheckResourceAttr("sonatypeiq_organization_role_membership.test", "role_id", "1cddabf7fdaa47d6833454af10e0a3ef"),
				),
			},
		},
	})
}
