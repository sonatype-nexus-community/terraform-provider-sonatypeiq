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

type ConfigUserTokenModel struct {
	ID                             types.String `tfsdk:"id"`
	UserTokenDefaultExpirationDays types.Int32  `tfsdk:"user_token_default_expiration_days"`
	LastUpdated                    types.String `tfsdk:"last_updated"`
}

func (m *ConfigUserTokenModel) MapFromApi(api *sonatypeiq.ApiUserTokenConfigurationDTO) {
	m.ID = types.StringValue(common.STATE_ID_USER_TOKEN_CONFIGURATION)
	m.UserTokenDefaultExpirationDays = types.Int32PointerValue(api.UserTokenDefaultExpirationDays)
}

func (m *ConfigUserTokenModel) MapToApi() *sonatypeiq.ApiUserTokenConfigurationDTO {
	api := sonatypeiq.NewApiUserTokenConfigurationDTO()
	api.UserTokenDefaultExpirationDays = m.UserTokenDefaultExpirationDays.ValueInt32Pointer()
	return api
}
