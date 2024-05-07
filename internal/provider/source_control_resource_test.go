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
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSourceControlApplicationResource(t *testing.T) {
	firstState := "true"
	secondState := "false"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceControlApplicationResource(firstState),
				Check: resource.ComposeTestCheckFunc(
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("sonatypeiq_source_control.test", "remediation_pull_requests_enabled", firstState),
						resource.TestCheckResourceAttr("sonatypeiq_source_control.test", "remediation_pull_requests_enabled", firstState),
						resource.TestCheckResourceAttr("sonatypeiq_source_control.test", "remediation_pull_requests_enabled", firstState),
					),
				),
			},
			{
				Config: testAccSourceControlApplicationResource(secondState),
				Check: resource.ComposeTestCheckFunc(
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("sonatypeiq_source_control.test", "remediation_pull_requests_enabled", secondState),
						resource.TestCheckResourceAttr("sonatypeiq_source_control.test", "remediation_pull_requests_enabled", secondState),
						resource.TestCheckResourceAttr("sonatypeiq_source_control.test", "remediation_pull_requests_enabled", secondState),
					),
				),
			},
		},
	})
}

func TestAccSourceControlOrganizationResource(t *testing.T) {
	firstState := "true"
	secondState := "false"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceControlOrganizationResource(firstState),
				Check: resource.ComposeTestCheckFunc(
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("sonatypeiq_source_control.test", "remediation_pull_requests_enabled", firstState),
						resource.TestCheckResourceAttr("sonatypeiq_source_control.test", "remediation_pull_requests_enabled", firstState),
						resource.TestCheckResourceAttr("sonatypeiq_source_control.test", "remediation_pull_requests_enabled", firstState),
					),
				),
			},
			{
				Config: testAccSourceControlOrganizationResource(secondState),
				Check: resource.ComposeTestCheckFunc(
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("sonatypeiq_source_control.test", "remediation_pull_requests_enabled", secondState),
						resource.TestCheckResourceAttr("sonatypeiq_source_control.test", "remediation_pull_requests_enabled", secondState),
						resource.TestCheckResourceAttr("sonatypeiq_source_control.test", "remediation_pull_requests_enabled", secondState),
					),
				),
			},
		},
	})
}

func testAccSourceControlApplicationResource(enabled string) string {
	return fmt.Sprintf(providerConfig+`
data "sonatypeiq_application" "app_by_public_id" {
  public_id = "sandbox-application"
}

resource "sonatypeiq_source_control" "test" {
  owner_type = "application"
  owner_id = data.sonatypeiq_application.app_by_public_id.id
  base_branch = "my-cool-branch"
  remediation_pull_requests_enabled = %s
  pull_request_commenting_enabled = %s
  source_control_evaluation_enabled = %s
  repository_url = "https://github.com/sonatype-nexus-community/terraform-provider-sonatypeiq.git"
}`, enabled, enabled, enabled)
}

func testAccSourceControlOrganizationResource(enabled string) string {
	return fmt.Sprintf(providerConfig+`
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

resource "sonatypeiq_source_control" "test" {
  owner_type = "organization"
  owner_id = data.sonatypeiq_organization.sandbox.id
  remediation_pull_requests_enabled = %s
  pull_request_commenting_enabled = %s
  source_control_evaluation_enabled = %s
  base_branch = "my-cool-branch"
}`, enabled, enabled, enabled)
}
