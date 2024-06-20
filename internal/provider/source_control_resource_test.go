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

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSourceControlApplicationResourceMinimumConfig(t *testing.T) {
	rand := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "sonatypeiq_source_control.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceControlApplicationResourceMinimumConfig(rand),
				Check: resource.ComposeTestCheckFunc(
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "repository_url", "https://github.com/sonatype-nexus-community/terraform-provider-sonatypeiq.git"),
						resource.TestCheckNoResourceAttr(resourceName, "remediation_pull_requests_enabled"),
						resource.TestCheckNoResourceAttr(resourceName, "pull_request_commenting_enabled"),
						resource.TestCheckNoResourceAttr(resourceName, "source_control_evaluation_enabled"),
						resource.TestCheckNoResourceAttr(resourceName, "base_branch"),
					),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "owner_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					id := s.RootModule().Resources[resourceName].Primary.Attributes["owner_id"]
					return fmt.Sprintf("application:%s", id), nil
				},
			},
		},
	})
}

func TestAccSourceControlApplicationResource(t *testing.T) {
	rand := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	firstState := "true"
	secondState := "false"
	resourceName := "sonatypeiq_source_control.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceControlApplicationResource(rand, firstState),
				Check: resource.ComposeTestCheckFunc(
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "remediation_pull_requests_enabled", firstState),
						resource.TestCheckResourceAttr(resourceName, "pull_request_commenting_enabled", firstState),
						resource.TestCheckResourceAttr(resourceName, "source_control_evaluation_enabled", firstState),
						resource.TestCheckResourceAttr(resourceName, "base_branch", "my-cool-branch"),
					),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "owner_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					id := s.RootModule().Resources[resourceName].Primary.Attributes["owner_id"]
					return fmt.Sprintf("application:%s", id), nil
				},
			},
			{
				Config: testAccSourceControlApplicationResource(rand, secondState),
				Check: resource.ComposeTestCheckFunc(
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "remediation_pull_requests_enabled", secondState),
						resource.TestCheckResourceAttr(resourceName, "pull_request_commenting_enabled", secondState),
						resource.TestCheckResourceAttr(resourceName, "source_control_evaluation_enabled", secondState),
						resource.TestCheckResourceAttr(resourceName, "base_branch", "my-cool-branch"),
					),
				),
			},
		},
	})
}

func TestAccSourceControlOrganizationResource(t *testing.T) {
	rand := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	firstState := "true"
	secondState := "false"
	resourceName := "sonatypeiq_source_control.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceControlOrganizationResource(rand, firstState),
				Check: resource.ComposeTestCheckFunc(
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "remediation_pull_requests_enabled", firstState),
						resource.TestCheckResourceAttr(resourceName, "pull_request_commenting_enabled", firstState),
						resource.TestCheckResourceAttr(resourceName, "source_control_evaluation_enabled", firstState),
						resource.TestCheckResourceAttr(resourceName, "base_branch", "my-cool-branch"),
					),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "owner_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					id := s.RootModule().Resources[resourceName].Primary.Attributes["owner_id"]
					return fmt.Sprintf("organization:%s", id), nil
				},
			},
			{
				Config: testAccSourceControlOrganizationResource(rand, secondState),
				Check: resource.ComposeTestCheckFunc(
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "remediation_pull_requests_enabled", secondState),
						resource.TestCheckResourceAttr(resourceName, "pull_request_commenting_enabled", secondState),
						resource.TestCheckResourceAttr(resourceName, "source_control_evaluation_enabled", secondState),
						resource.TestCheckResourceAttr(resourceName, "base_branch", "my-cool-branch"),
					),
				),
			},
		},
	})
}

func testAccSourceControlApplicationResource(rand string, enabled string) string {
	return fmt.Sprintf(providerConfig+`
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

resource "sonatypeiq_application" "app_by_public_id" {
  name = "app-%s"
  public_id = "app-%s"
  organization_id = data.sonatypeiq_organization.sandbox.id
}

resource "sonatypeiq_source_control" "test" {
  owner_type = "application"
  owner_id = sonatypeiq_application.app_by_public_id.id
  base_branch = "my-cool-branch"
  remediation_pull_requests_enabled = %s
  pull_request_commenting_enabled = %s
  source_control_evaluation_enabled = %s
  repository_url = "https://github.com/sonatype-nexus-community/terraform-provider-sonatypeiq.git"
}`, rand, rand, enabled, enabled, enabled)
}

func testAccSourceControlApplicationResourceMinimumConfig(rand string) string {
	return fmt.Sprintf(providerConfig+`
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

resource "sonatypeiq_application" "app_by_public_id" {
  name = "app-%s"
  public_id = "app-%s"
  organization_id = data.sonatypeiq_organization.sandbox.id
}

resource "sonatypeiq_source_control" "test" {
  owner_type = "application"
  owner_id = sonatypeiq_application.app_by_public_id.id
  repository_url = "https://github.com/sonatype-nexus-community/terraform-provider-sonatypeiq.git"
}`, rand, rand)
}

func testAccSourceControlOrganizationResource(rand string, enabled string) string {
	return fmt.Sprintf(providerConfig+`
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

resource "sonatypeiq_organization" "my_sandbox" {
  name = "Sandbox Organization %s"
  parent_organization_id = data.sonatypeiq_organization.sandbox.id
}

resource "sonatypeiq_source_control" "test" {
  owner_type = "organization"
  owner_id = sonatypeiq_organization.my_sandbox.id
  remediation_pull_requests_enabled = %s
  pull_request_commenting_enabled = %s
  source_control_evaluation_enabled = %s
  base_branch = "my-cool-branch"
}`, rand, enabled, enabled, enabled)
}
