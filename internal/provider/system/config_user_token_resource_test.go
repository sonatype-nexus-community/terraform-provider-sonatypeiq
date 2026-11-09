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
	"terraform-provider-sonatypeiq/internal/provider/testutil"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfigUserTokenResource(t *testing.T) {
	// The endpoint was introduced in release 199
	testutil.SkipIfNxiqVersionOlderThan(t, 199)

	resourceName := "sonatypeiq_user_token_configuration.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccConfigUserTokenResource(90),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_USER_TOKEN_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "user_token_default_expiration_days", "90"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Update testing
			{
				Config: testAccConfigUserTokenResource(45),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_USER_TOKEN_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "user_token_default_expiration_days", "45"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate no drift
			{
				Config:             testAccConfigUserTokenResource(45),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Import testing
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

func testAccConfigUserTokenResource(expirationDays int) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_user_token_configuration" "test" {
  user_token_default_expiration_days = %d
}`, expirationDays)
}
