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

// SystemConfigModel
// ------------------------------------------------------------
type SystemConfigModel struct {
	ID           types.String `tfsdk:"id"`
	BaseURL      types.String `tfsdk:"base_url"`
	ForceBaseURL types.Bool   `tfsdk:"force_base_url"`
}

func (m *SystemConfigModel) MapFromApi(api *sonatypeiq.SystemConfig) {
	m.BaseURL = types.StringPointerValue(api.BaseUrl.Get())
	m.ForceBaseURL = types.BoolPointerValue(api.ForceBaseUrl.Get())
}

// SystemConfigResource
// ------------------------------------------------------------
type SystemConfigResource struct {
	SystemConfigModel
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (m *SystemConfigResource) MapToApi() *sonatypeiq.SystemConfig {
	api := sonatypeiq.NewSystemConfigWithDefaults()
	if !m.BaseURL.IsNull() {
		api.BaseUrl = *sonatypeiq.NewNullableString(m.BaseURL.ValueStringPointer())
	}
	if !m.ForceBaseURL.IsNull() {
		api.ForceBaseUrl = *sonatypeiq.NewNullableBool(m.ForceBaseURL.ValueBoolPointer())
	}
	return api
}

func (m *SystemConfigResource) MapFromApi(api *sonatypeiq.SystemConfig) {
	m.ID = types.StringValue(common.STATE_ID_SYSTEM_CONFIGURATION)
	m.BaseURL = types.StringPointerValue(api.BaseUrl.Get())
	m.ForceBaseURL = types.BoolPointerValue(api.ForceBaseUrl.Get())
}
