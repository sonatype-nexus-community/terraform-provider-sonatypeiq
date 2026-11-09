package system_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
)

func TestUserTokenConfigurationResource_VersionGating_Offline(t *testing.T) {
	// On a server older than release 199, the resource must be refused at plan time
	t.Run("older_server_refuses", func(t *testing.T) {
		newFakeUserTokenConfigIqServer(t, 195)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      userTokenConfig(90),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`user_token_configuration resource was introduced in Sonatype IQ Server\s+release 199, but the connected server is release 195`),
				},
			},
		})
	})

	// On a server at exactly release 199, the resource is accepted
	t.Run("exactly_198_works", func(t *testing.T) {
		f := newFakeUserTokenConfigIqServer(t, 199)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: userTokenConfig(90),
					Check:  f.checkStored(int32Ptr(90)),
				},
			},
		})
	})

	// A server whose version cannot be determined (release 0) is not blocked
	t.Run("undetected_version_not_blocked", func(t *testing.T) {
		f := newFakeUserTokenConfigIqServer(t, 0) // version 0 means undetected
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: userTokenConfig(90),
					Check:  f.checkStored(int32Ptr(90)),
				},
			},
		})
	})
}

func TestUserTokenConfigurationResource_RequiredAttribute_Offline(t *testing.T) {
	newFakeUserTokenConfigIqServer(t, 199)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `provider "sonatypeiq" {
  username = ""
  password = ""
  url      = ""
}

resource "sonatypeiq_user_token_configuration" "test" {
}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`The argument "user_token_default_expiration_days" is required`),
			},
		},
	})
}

func int32Ptr(i int32) *int32 {
	return &i
}
