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

package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

type SecuritySamlModel struct {
	IdentityProviderName       types.String `tfsdk:"identity_provider_name"`
	IdpMetadata                types.String `tfsdk:"idp_metadata"`
	UsernameAttribute          types.String `tfsdk:"username_attribute"`
	FirstNameAttribute         types.String `tfsdk:"first_name_attribute"`
	LastNameAttribute          types.String `tfsdk:"last_name_attribute"`
	EmailAttribute             types.String `tfsdk:"email_attribute"`
	GroupsAttribute            types.String `tfsdk:"groups_attribute"`
	ValidateResponseSignature  types.Bool   `tfsdk:"validate_response_signature"`
	ValidateAssertionSignature types.Bool   `tfsdk:"validate_assertion_signature"`
	EntityId                   types.String `tfsdk:"entity_id"`
}

func (m *SecuritySamlModel) MapToApi(api *sonatypeiq.ApiSamlConfigurationDTO) {
	api.IdentityProviderName = m.IdentityProviderName.ValueStringPointer()
	api.EntityId = m.EntityId.ValueStringPointer()
	api.FirstNameAttributeName = m.FirstNameAttribute.ValueStringPointer()
	api.LastNameAttributeName = m.LastNameAttribute.ValueStringPointer()
	api.EmailAttributeName = m.EmailAttribute.ValueStringPointer()
	api.UsernameAttributeName = m.UsernameAttribute.ValueStringPointer()
	api.GroupsAttributeName = m.GroupsAttribute.ValueStringPointer()
	api.ValidateAssertionSignature = m.ValidateAssertionSignature.ValueBoolPointer()
	api.ValidateResponseSignature = m.ValidateResponseSignature.ValueBoolPointer()
}

func (m *SecuritySamlModel) MapFromApi(api *sonatypeiq.ApiSamlConfigurationResponseDTO) {
	m.IdentityProviderName = types.StringPointerValue(api.IdentityProviderName)
	m.IdpMetadata = types.StringPointerValue(api.IdentityProviderMetadataXml)
	m.UsernameAttribute = types.StringPointerValue(api.UsernameAttributeName)
	m.FirstNameAttribute = types.StringPointerValue(api.FirstNameAttributeName)
	m.LastNameAttribute = types.StringPointerValue(api.LastNameAttributeName)
	m.EmailAttribute = types.StringPointerValue(api.EmailAttributeName)
	m.GroupsAttribute = types.StringPointerValue(api.GroupsAttributeName)
	m.ValidateResponseSignature = types.BoolPointerValue(api.ValidateResponseSignature)
	m.ValidateAssertionSignature = types.BoolPointerValue(api.ValidateAssertionSignature)
	m.EntityId = types.StringPointerValue(api.EntityId)
}
