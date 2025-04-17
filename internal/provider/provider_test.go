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
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration.
	providerConfig = `
provider "sonatypeiq" {}
`
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"sonatypeiq": providerserver.NewProtocol6WithError(New("test")()),
}

//	func testAccPreCheck(t *testing.T) {
//		// You can add code here to run prior to any test case execution, for example assertions
//		// about the appropriate environment variables being set are common to see in a pre-check
//		// function.
//	}

func TestAccProviderNoConfigurationEnvVarsEmpty(t *testing.T) {
	originalUsername, usernameSet := os.LookupEnv("IQ_SERVER_USERNAME")
	_ = os.Unsetenv("IQ_SERVER_USERNAME")
	originalPassword, PasswordSet := os.LookupEnv("IQ_SERVER_PASSWORD")
	_ = os.Unsetenv("IQ_SERVER_PASSWORD")
	originalServerURL, serverURLSet := os.LookupEnv("IQ_SERVER_URL")
	_ = os.Unsetenv("IQ_SERVER_URL")
	defer func() {
		if usernameSet {
			_ = os.Setenv("IQ_SERVER_USERNAME", originalUsername)
		}
		if PasswordSet {
			_ = os.Setenv("IQ_SERVER_PASSWORD", originalPassword)
		}
		if serverURLSet {
			_ = os.Setenv("IQ_SERVER_URL", originalServerURL)
		}
	}()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `resource "sonatypeiq_application" "test" {
  name = "test"
  public_id = "test"
  organization_id = "aaaaa"
}`,
				ExpectError: regexp.MustCompile("(?s).*Unknown Sonatype IQ Server URL.*Invalid Sonatype IQ Server URL.*Username not supplied.*password not supplied.*"),
			},
		},
	})
}
