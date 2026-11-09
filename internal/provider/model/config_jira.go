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
	"encoding/json"
	"reflect"
	"terraform-provider-sonatypeiq/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

type ConfigJiraModel struct {
	ID           types.String `tfsdk:"id"`
	Url          types.String `tfsdk:"url"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	CustomFields types.String `tfsdk:"custom_fields"`
	LastUpdated  types.String `tfsdk:"last_updated"`
}

func (m *ConfigJiraModel) MapFromApi(api *sonatypeiq.ApiJiraConfigurationDTO) error {
	m.ID = types.StringValue(common.STATE_ID_JIRA_CONFIGURATION)
	m.Url = types.StringPointerValue(api.Url)
	m.Username = types.StringPointerValue(api.Username)
	// The password is never returned by the API, so the value in state is kept.

	if len(api.CustomFields) == 0 {
		m.CustomFields = types.StringNull()
		return nil
	}

	// The server normalises the JSON (sorted keys, no white space): keep the user's own string when it means the same
	if customFieldsEqual(m.CustomFields, api.CustomFields) {
		return nil
	}

	customFieldsJSON, err := json.Marshal(api.CustomFields)
	if err != nil {
		return err
	}
	m.CustomFields = types.StringValue(string(customFieldsJSON))
	return nil
}

// customFieldsEqual reports whether the JSON string held in state describes the same custom fields as the server.
func customFieldsEqual(current types.String, fromApi map[string]map[string]interface{}) bool {
	if current.IsNull() || current.IsUnknown() {
		return false
	}

	var currentFields map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(current.ValueString()), &currentFields); err != nil {
		return false
	}

	// Round trip the server's value so numbers and nested values have the same Go types on both sides
	apiJSON, err := json.Marshal(fromApi)
	if err != nil {
		return false
	}
	var apiFields map[string]map[string]interface{}
	if err := json.Unmarshal(apiJSON, &apiFields); err != nil {
		return false
	}

	return reflect.DeepEqual(currentFields, apiFields)
}

func (m *ConfigJiraModel) MapToApi() (*sonatypeiq.ApiJiraConfigurationDTO, error) {
	api := sonatypeiq.NewApiJiraConfigurationDTOWithDefaults()
	api.Url = m.Url.ValueStringPointer()
	api.Username = m.Username.ValueStringPointer()
	api.Password = m.Password.ValueStringPointer()

	// Parse customFields from JSON string if set
	if !m.CustomFields.IsNull() && m.CustomFields.ValueString() != "" {
		var customFields map[string]map[string]interface{}
		if err := json.Unmarshal([]byte(m.CustomFields.ValueString()), &customFields); err != nil {
			return nil, err
		}
		api.CustomFields = customFields
	}

	return api, nil
}
