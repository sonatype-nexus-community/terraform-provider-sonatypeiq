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

type ConfigCrowdModel struct {
	ServerUrl           types.String `tfsdk:"server_url"`
	ApplicationName     types.String `tfsdk:"application_name"`
	ApplicationPassword types.String `tfsdk:"application_password"`
	LastUpdated         types.String `tfsdk:"last_updated"`
}

func (m *ConfigCrowdModel) MapFromApi(api *sonatypeiq.ApiCrowdConfigurationDTO) {
	m.ServerUrl = sharedutil.StringPtrToValue(api.ServerUrl)
	m.ApplicationName = sharedutil.StringPtrToValue(api.ApplicationName)
}

func (m *ConfigCrowdModel) MapToApi(api *sonatypeiq.ApiCrowdConfigurationDTO) {
	api.ServerUrl = sharedutil.StringToPtr(m.ServerUrl.ValueString())
	api.ApplicationName = sharedutil.StringToPtr(m.ApplicationName.ValueString())
	api.ApplicationPassword = sharedutil.StringToPtr(m.ApplicationPassword.ValueString())
}
