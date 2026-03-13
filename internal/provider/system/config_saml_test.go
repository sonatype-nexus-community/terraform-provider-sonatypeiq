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
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-sonatypeiq/internal/provider/common"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
)

func TestAccConfigSamlResource(t *testing.T) {
	randomStr := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "sonatypeiq_config_saml.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSecuritySamlResourceConfigMinimum(randomStr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_SAML_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_name", fmt.Sprintf("Test-IDP-%s", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "username_attribute", "username"),
					resource.TestCheckResourceAttr(resourceName, "first_name_attribute", common.SAML_DEFAULT_FIRST_NAME_ATTRIBUTE),
					resource.TestCheckResourceAttr(resourceName, "last_name_attribute", common.SAML_DEFAULT_LAST_NAME_ATTRIBUTE),
					resource.TestCheckResourceAttr(resourceName, "email_attribute", common.SAML_DEFAULT_EMAIL_ATTRIBUTE),
					resource.TestCheckResourceAttr(resourceName, "groups_attribute", common.SAML_DEFAULT_GROUPS_ATTRIBUTE),
					resource.TestCheckResourceAttr(resourceName, "validate_response_signature", "false"),
					resource.TestCheckResourceAttr(resourceName, "validate_assertion_signature", "false"),
					resource.TestCheckResourceAttr(resourceName, "entity_id", fmt.Sprintf("test-entity-%s", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "idp_metadata", testSamlMetadata(false)),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			{
				Config: testAccSecuritySamlResourceConfigFull(randomStr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", common.STATE_ID_SAML_CONFIGURATION),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_name", fmt.Sprintf("Test-IDP-%s", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "username_attribute", "updatedUsername"),
					resource.TestCheckResourceAttr(resourceName, "first_name_attribute", "updatedFirstName"),
					resource.TestCheckResourceAttr(resourceName, "last_name_attribute", "updatedLastName"),
					resource.TestCheckResourceAttr(resourceName, "email_attribute", "updatedEmail"),
					resource.TestCheckResourceAttr(resourceName, "groups_attribute", "updatedGroups"),
					resource.TestCheckResourceAttr(resourceName, "validate_response_signature", "true"),
					resource.TestCheckResourceAttr(resourceName, "validate_assertion_signature", "true"),
					resource.TestCheckResourceAttr(resourceName, "entity_id", fmt.Sprintf("updated-entity-%s", randomStr)),
					resource.TestCheckResourceAttr(resourceName, "idp_metadata", testSamlMetadata(false)),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Validate
			{
				Config:             testAccSecuritySamlResourceConfigFull(randomStr),
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

func testAccSecuritySamlResourceConfigMinimum(randomSuffix string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_config_saml" "test" {
  identity_provider_name = "Test-IDP-%s"
  idp_metadata = %s
  username_attribute = "username"
  entity_id = "test-entity-%s"
}
`, randomSuffix, testSamlMetadata(true), randomSuffix)
}

func testAccSecuritySamlResourceConfigFull(randomSuffix string) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_config_saml" "test" {
  identity_provider_name = "Test-IDP-%s"
  idp_metadata = %s
  username_attribute = "updatedUsername"
  first_name_attribute = "updatedFirstName"
  last_name_attribute = "updatedLastName"
  email_attribute = "updatedEmail"
  groups_attribute = "updatedGroups"
  validate_response_signature = true
  validate_assertion_signature = true
  entity_id = "updated-entity-%s"
}
`, randomSuffix, testSamlMetadata(true), randomSuffix)
}

func testSamlMetadata(includeHeredoc bool) string {
	metadata := `<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" ID="test-id" entityID="https://test.example.com/saml">
  <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <KeyDescriptor use="signing">
      <KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#">
        <X509Data>
          <X509Certificate>
MIIEETCCAvmgAwIBAgIUB85DwdgvP7Lp6hIv4R6F84p3H3wwDQYJKoZIhvcNAQEL
BQAwgZcxCzAJBgNVBAYTAkFVMS4wLAYDVQQIDCUjIFNoYXJlIG9ubHkgdGhlIGNl
cnRpZmljYXRlIHdpdGggSWRQMTUwMwYDVQQHDCwjIEtlZXAgdGhlIHByaXZhdGUg
a2V5IHNlY3VyZSBvbiB5b3VyIHNlcnZlcjEhMB8GA1UECgwYSW50ZXJuZXQgV2lk
Z2l0cyBQdHkgTHRkMB4XDTI1MDkxNDIzNDgwN1oXDTI2MDkxNDIzNDgwN1owgZcx
CzAJBgNVBAYTAkFVMS4wLAYDVQQIDCUjIFNoYXJlIG9ubHkgdGhlIGNlcnRpZmlj
YXRlIHdpdGggSWRQMTUwMwYDVQQHDCwjIEtlZXAgdGhlIHByaXZhdGUga2V5IHNl
Y3VyZSBvbiB5b3VyIHNlcnZlcjEhMB8GA1UECgwYSW50ZXJuZXQgV2lkZ2l0cyBQ
dHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1nIHufmHKbyj
PNDb/KaZ6ppXuMdy87/yXsTikH9dDQe5Ya/reihsfGNY3+uEOhRygKiqasHWkseZ
YLn1rLnhDAwGp/8MTcLYqY3S0F2eF/1q1SjiRJz9++svMLFwoMLHie64zqJJINqA
B0Vp0hu7HVCU8ZSKPBfS6PW5w5doO9huCCyRviabojG9RfqSk5VkQLG3TsjpMwaI
UzeLlMC9GTqmW3moUwkwoMnbKT4YCq/tNHwpDGq0Du1y7B2o+0Z+pihCLN5WYEZT
jiLPvGOpzQcSI3n8Te+qoU7J1qQ6c1BJhs2GUF2cOhLQIvLgHDNB0v94Xfz3ma5Y
ft5REQAHwQIDAQABo1MwUTAdBgNVHQ4EFgQUWwVsy9QgHO/bS81VCkINE9UujPMw
HwYDVR0jBBgwFoAUWwVsy9QgHO/bS81VCkINE9UujPMwDwYDVR0TAQH/BAUwAwEB
/zANBgkqhkiG9w0BAQsFAAOCAQEASjabAHp0KqyyyTSVta6x4QAL5Vxg9vcTAklK
iXjvMOzExXYm1zI9eOvIzj1fn0e0tLOA03LGLHkLdQK9A5x/MqXysmK8sXsy5JBv
5WwvFE/7kKefaGsHSOLiXgTpgJtf3X0Z3B4s8y3Y2i6KvPA6uGElNbtd7WvTTFwU
oMhYx6Dkmek4o31fqmvuo3JbOPz2+zJOcD1Y3qL7He7WaSBLA9s3CQwYcTPBjy6i
HYoEGvzPl9flJfVRQ/GdVQxYZTAtmOgy+nSZwkm4oz2Xoz3q8/kHoOIBuPB6iWzf
XBWQDGEjEr6wHCaLl3CLN3yIKxXzJW2wqFvpm0MNcM0E9W9gXQ==
          </X509Certificate>
        </X509Data>
      </KeyInfo>
    </KeyDescriptor>
    <SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://test.example.com/saml/logout"/>
    <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://test.example.com/saml/sso"/>
    <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://test.example.com/saml/sso"/>
  </IDPSSODescriptor>
</EntityDescriptor>`

	if includeHeredoc {
		return fmt.Sprintf(`<<-EOT
%s
EOT
`, metadata)
	} else {
		return metadata + "\n"
	}
}
