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
	"terraform-provider-sonatypeiq/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// serverDefaultUsernameHeader is the header Sonatype IQ Server uses when none is configured.
const serverDefaultUsernameHeader = "REMOTE_USER"

type ConfigReverseProxyAuthenticationModel struct {
	ID                     types.String `tfsdk:"id"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	UsernameHeader         types.String `tfsdk:"username_header"`
	CsrfProtectionDisabled types.Bool   `tfsdk:"csrf_protection_disabled"`
	LogoutUrl              types.String `tfsdk:"logout_url"`
	LastUpdated            types.String `tfsdk:"last_updated"`
}

func (m *ConfigReverseProxyAuthenticationModel) MapFromApi(api *sonatypeiq.ApiReverseProxyAuthenticationConfigurationDTO, state *ConfigReverseProxyAuthenticationModel) {
	m.ID = types.StringValue(common.STATE_ID_REVERSE_PROXY_AUTHENTICATION_CONFIGURATION)
	m.Enabled = types.BoolPointerValue(api.Enabled)

	// Detect import: when LastUpdated is null (ID is set by ImportStatePassthroughID,
	// but LastUpdated is not set until after create/update)
	isImport := state == nil || state.LastUpdated.IsNull()

	// Preserve null for optional attributes that were not set in state, EXCEPT:
	// - during import (accept all values from API)
	// - when server has a non-default value (that would be drift)
	// For usernameHeader, the server default is serverDefaultUsernameHeader
	if !isImport && state.UsernameHeader.IsNull() {
		if api.UsernameHeader != nil && *api.UsernameHeader != serverDefaultUsernameHeader {
			// Server has a non-default value - this is drift, don't preserve null
			m.UsernameHeader = types.StringPointerValue(api.UsernameHeader)
		} else {
			m.UsernameHeader = types.StringNull()
		}
	} else {
		m.UsernameHeader = types.StringPointerValue(api.UsernameHeader)
	}

	if !isImport && state.CsrfProtectionDisabled.IsNull() {
		m.CsrfProtectionDisabled = types.BoolNull()
	} else {
		m.CsrfProtectionDisabled = types.BoolPointerValue(api.CsrfProtectionDisabled)
	}

	if !isImport && state.LogoutUrl.IsNull() {
		m.LogoutUrl = types.StringNull()
	} else {
		m.LogoutUrl = types.StringPointerValue(api.LogoutUrl)
	}
}

func (m *ConfigReverseProxyAuthenticationModel) MapToApi() *sonatypeiq.ApiReverseProxyAuthenticationConfigurationDTO {
	api := sonatypeiq.NewApiReverseProxyAuthenticationConfigurationDTOWithDefaults()
	api.Enabled = m.Enabled.ValueBoolPointer()
	api.UsernameHeader = m.UsernameHeader.ValueStringPointer()
	api.CsrfProtectionDisabled = m.CsrfProtectionDisabled.ValueBoolPointer()
	api.LogoutUrl = m.LogoutUrl.ValueStringPointer()
	return api
}
