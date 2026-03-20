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

type UserTokenModel struct {
	GeneratedAt types.String `tfsdk:"generated_at"`
	// Username    types.String `tfsdk:"username"`
	// Realm       types.String `tfsdk:"realm"`
	UserCode types.String `tfsdk:"user_code"`
	PassCode types.String `tfsdk:"pass_code"`
}

func (m *UserTokenModel) MapFromApi(api *sonatypeiq.ApiUserTokenDTO) {
	// m.Username = sharedutil.StringPtrToValue(api.Username)
	// m.Realm = sharedutil.StringPtrToValue(api.Realm)
	m.UserCode = sharedutil.StringPtrToValue(api.UserCode)
	m.PassCode = sharedutil.StringPtrToValue(api.PassCode)
}

type UserTokenModelResource struct {
	UserTokenModel
	LastUpdated types.String `tfsdk:"last_updated"`
}
