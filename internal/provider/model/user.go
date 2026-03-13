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
	"fmt"
	"terraform-provider-sonatypeiq/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// UserModel
// ------------------------------------------------------------
type UserModel struct {
	ID          types.String `tfsdk:"id"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	FirstName   types.String `tfsdk:"first_name"`
	LastName    types.String `tfsdk:"last_name"`
	Email       types.String `tfsdk:"email"`
	Realm       types.String `tfsdk:"realm"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (m *UserModel) MapToApi(includePassword bool) *sonatypeiq.ApiUserDTO {
	api := sonatypeiq.NewApiUserDTOWithDefaults()
	api.Username = m.Username.ValueStringPointer()
	if includePassword {
		api.Password = m.Password.ValueStringPointer()
	}
	api.FirstName = m.FirstName.ValueStringPointer()
	api.LastName = m.LastName.ValueStringPointer()
	api.Email = m.Email.ValueStringPointer()
	api.Realm = m.Realm.ValueStringPointer()
	return api
}

func (m *UserModel) MapFromApi(api *sonatypeiq.ApiUserDTO) {
	m.ID = types.StringValue(fmt.Sprintf(common.USER_ID_FORMAT, api.GetRealm(), api.GetUsername()))
	m.Username = types.StringPointerValue(api.Username)
	// Password not returned by API
	m.FirstName = types.StringPointerValue(api.FirstName)
	m.LastName = types.StringPointerValue(api.LastName)
	m.Email = types.StringPointerValue(api.Email)
	m.Realm = types.StringPointerValue(api.Realm)
}
