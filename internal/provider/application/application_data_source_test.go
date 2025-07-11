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

package application_test

import (
	"testing"

	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: utils_test.ProviderConfig + `data "sonatypeiq_application" "app_by_public_id" {
					public_id = "sandbox-application"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sonatypeiq_application.app_by_public_id", "id"),
					resource.TestCheckResourceAttr("data.sonatypeiq_application.app_by_public_id", "public_id", "sandbox-application"),
					resource.TestCheckResourceAttr("data.sonatypeiq_application.app_by_public_id", "name", "Sandbox Application"),
					resource.TestCheckResourceAttrSet("data.sonatypeiq_application.app_by_public_id", "organization_id"),
				),
			},
		},
	})
}
