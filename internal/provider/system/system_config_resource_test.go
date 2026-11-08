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
	"os"
	"regexp"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/testutil"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The release that introduced the most recent System Configuration property (integrations_supported_version_count)
const systemConfigAllPropertiesRelease = 197

func TestAccSystemConfigResource(t *testing.T) {
	iqUrl := os.Getenv("IQ_SERVER_URL") + "/"
	resourceName := "sonatypeiq_system_config.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSystemConfigResource(iqUrl, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_SYSTEM_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "base_url", iqUrl),
					resource.TestCheckResourceAttr(resourceName, "force_base_url", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			{
				Config: testAccSystemConfigResource(iqUrl, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_SYSTEM_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "base_url", iqUrl),
					resource.TestCheckResourceAttr(resourceName, "force_base_url", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccSystemConfigResource(iqUrl, false),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
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

func testAccSystemConfigResource(url string, force bool) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_system_config" "test" {
  base_url = "%s"
  force_base_url = %t
}`, url, force)
}

// Every System Configuration property exists from release 197 onwards.
func TestAccSystemConfigResource_AllProperties(t *testing.T) {
	if v := testutil.DetectedIqVersion(t); v < systemConfigAllPropertiesRelease {
		t.Skipf("Sonatype IQ Server %d does not support every System Configuration property (needs %d)", v, systemConfigAllPropertiesRelease)
	}

	iqUrl := os.Getenv("IQ_SERVER_URL") + "/"
	resourceName := "sonatypeiq_system_config.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemConfigResourceAllProperties(iqUrl, 45, `user_agent_suffix = "Approved"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "session_timeout", "45"),
					resource.TestCheckResourceAttr(resourceName, "csrf_protection", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy_monitoring_hour", "3"),
					resource.TestCheckResourceAttr(resourceName, "max_advanced_search_clause_count", "1024"),
					resource.TestCheckResourceAttr(resourceName, "advanced_search_csv_export_delimiter", ";"),
					resource.TestCheckResourceAttr(resourceName, "success_metrics_stage_id", "build"),
					resource.TestCheckResourceAttr(resourceName, "integrations_supported_version_count", "5"),
					resource.TestCheckResourceAttr(resourceName, "api_access_allow_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "frame_ancestors_allowlist.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_agent_suffix", "Approved"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Change a property, and remove one: the removed property must be unmanaged afterwards
			{
				Config: testAccSystemConfigResourceAllProperties(iqUrl, 60, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "session_timeout", "60"),
					resource.TestCheckNoResourceAttr(resourceName, "user_agent_suffix"),
				),
			},
			// No drift
			{
				Config:             testAccSystemConfigResourceAllProperties(iqUrl, 60, ""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// Setting a property that is newer than the connected Sonatype IQ Server must fail at plan time and explain why.
func TestAccSystemConfigResource_UnsupportedProperty(t *testing.T) {
	v := testutil.DetectedIqVersion(t)
	if v >= systemConfigAllPropertiesRelease {
		t.Skipf("Sonatype IQ Server %d supports every System Configuration property", v)
	}

	iqUrl := os.Getenv("IQ_SERVER_URL") + "/"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_system_config" "test" {
  base_url                             = "%s"
  force_base_url                       = false
  integrations_supported_version_count = 5
}`, iqUrl),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`introduced in Sonatype IQ Server\s+release\s+%d`, systemConfigAllPropertiesRelease)),
			},
		},
	})
}

func testAccSystemConfigResourceAllProperties(url string, sessionTimeout int, extra string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_system_config" "test" {
  base_url                                    = "%s"
  force_base_url                              = false
  session_timeout                             = %d
  csrf_protection                             = true
  policy_monitoring_hour                      = 3
  max_advanced_search_clause_count            = 1024
  advanced_search_csv_export_delimiter        = ";"
  success_metrics_stage_id                    = "build"
  integrations_supported_version_count        = 5
  waived_component_upgrade_monitoring_enabled = false
  alp_observed_license_detection_enabled      = true
  advanced_reporting_insights_enabled         = true
  api_access_allow_list                       = ["admin"]
  frame_ancestors_allowlist                   = ["*.example.com"]
  %s
}`, url, sessionTimeout, extra)
}
