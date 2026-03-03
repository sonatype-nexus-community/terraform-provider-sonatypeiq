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
	sharedutil "github.com/sonatype-nexus-community/terraform-provider-shared/util"
)

// SecuritySamlMetadataModel
// ------------------------------------------------------------
type SecuritySamlMetadataModel struct {
	ID           types.String `tfsdk:"id"`
	SamlMetadata types.String `tfsdk:"saml_metadata"`
}

// SecuritySamlModel
// ------------------------------------------------------------
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
	LastUpdated                types.String `tfsdk:"last_updated"`
}

func (m *SecuritySamlModel) MapToApi(api *sonatypeiq.ApiSamlConfigurationDTO) {
	api.IdentityProviderName = sharedutil.StringToPtr(m.IdentityProviderName.ValueString())
	api.EntityId = sharedutil.StringToPtr(m.EntityId.ValueString())
	api.FirstNameAttributeName = sharedutil.StringToPtr(m.FirstNameAttribute.ValueString())
	api.LastNameAttributeName = sharedutil.StringToPtr(m.LastNameAttribute.ValueString())
	api.EmailAttributeName = sharedutil.StringToPtr(m.EmailAttribute.ValueString())
	api.UsernameAttributeName = sharedutil.StringToPtr(m.UsernameAttribute.ValueString())
	api.GroupsAttributeName = sharedutil.StringToPtr(m.GroupsAttribute.ValueString())
	api.ValidateAssertionSignature = sharedutil.BoolToPtr(m.ValidateAssertionSignature.ValueBool())
	api.ValidateResponseSignature = sharedutil.BoolToPtr(m.ValidateResponseSignature.ValueBool())
}

func (m *SecuritySamlModel) MapFromApi(api *sonatypeiq.ApiSamlConfigurationResponseDTO) {
	m.IdentityProviderName = sharedutil.StringPtrToValue(api.IdentityProviderName)
	m.IdpMetadata = sharedutil.StringPtrToValue(api.IdentityProviderMetadataXml)
	m.UsernameAttribute = sharedutil.StringPtrToValue(api.UsernameAttributeName)
	m.FirstNameAttribute = sharedutil.StringPtrToValue(api.FirstNameAttributeName)
	m.LastNameAttribute = sharedutil.StringPtrToValue(api.LastNameAttributeName)
	m.EmailAttribute = sharedutil.StringPtrToValue(api.EmailAttributeName)
	m.GroupsAttribute = sharedutil.StringPtrToValue(api.GroupsAttributeName)
	m.ValidateResponseSignature = sharedutil.BoolPtrToValue(api.ValidateResponseSignature)
	m.ValidateAssertionSignature = sharedutil.BoolPtrToValue(api.ValidateAssertionSignature)
	m.EntityId = sharedutil.StringPtrToValue(api.EntityId)
}
