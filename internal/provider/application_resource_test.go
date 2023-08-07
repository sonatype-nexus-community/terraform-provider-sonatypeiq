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

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

resource "sonatypeiq_application" "test" {
  name = "My Example Application"
  public_id = "my_example_application"
  organization_id = data.sonatypeiq_organization.sandbox.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify Application
					resource.TestCheckResourceAttrSet("sonatypeiq_application.test", "id"),
					resource.TestCheckResourceAttr("sonatypeiq_application.test", "name", "My Example Application"),
					resource.TestCheckResourceAttr("sonatypeiq_application.test", "public_id", "my_example_application"),
					resource.TestCheckResourceAttrSet("sonatypeiq_application.test", "last_updated"),
				),
			},
			{
				Config: providerConfig + `
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

resource "sonatypeiq_application" "test" {
  name = "My Example Application 2"
  public_id = "my_example_application_2"
  organization_id = data.sonatypeiq_organization.sandbox.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sonatypeiq_application.test", "id"),
					resource.TestCheckResourceAttr("sonatypeiq_application.test", "name", "My Example Application 2"),
					resource.TestCheckResourceAttr("sonatypeiq_application.test", "public_id", "my_example_application_2"),
					resource.TestCheckResourceAttrSet("sonatypeiq_application.test", "last_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
