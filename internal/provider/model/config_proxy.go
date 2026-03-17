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
	"context"
	"terraform-provider-sonatypeiq/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

type ConfigProxyModel struct {
	ID           types.String `tfsdk:"id"`
	Hostname     types.String `tfsdk:"hostname"`
	Port         types.Int32  `tfsdk:"port"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	ExcludeHosts types.Set    `tfsdk:"exclude_hosts"`
	LastUpdated  types.String `tfsdk:"last_updated"`
}

func (m *ConfigProxyModel) MapFromApi(ctx context.Context, api *sonatypeiq.ApiProxyServerConfigurationDTO) {
	m.ID = types.StringValue(common.STATE_ID_PROXY_CONFIGURATION)
	m.Hostname = types.StringPointerValue(api.Hostname)
	m.Port = types.Int32PointerValue(api.Port)
	m.Username = types.StringPointerValue(api.Username)
	// Password never returned by API
	excludeHosts := make([]string, 0)
	for _, excludeHost := range api.GetExcludeHosts() {
		excludeHosts = append(excludeHosts, excludeHost)
	}
	m.ExcludeHosts, _ = types.SetValueFrom(ctx, types.StringType, excludeHosts)
}

func (m *ConfigProxyModel) MapToApi(ctx context.Context) *sonatypeiq.ApiProxyServerConfigurationDTO {
	api := sonatypeiq.NewApiProxyServerConfigurationDTOWithDefaults()
	api.Hostname = m.Hostname.ValueStringPointer()
	api.Port = m.Port.ValueInt32Pointer()
	api.Username = m.Username.ValueStringPointer()
	api.Password = m.Password.ValueStringPointer()
	if m.Password.IsNull() {
		api.PasswordIsIncluded = sonatypeiq.PtrBool(false)
	} else {
		api.PasswordIsIncluded = sonatypeiq.PtrBool(true)
	}
	var excludeHosts []string
	m.ExcludeHosts.ElementsAs(ctx, &excludeHosts, false)
	api.SetExcludeHosts(excludeHosts)
	return api
}
