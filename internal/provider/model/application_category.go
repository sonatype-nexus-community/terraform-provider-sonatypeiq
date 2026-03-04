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

// TagModel
// ------------------------------------------------------------
type TagModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Color       types.String `tfsdk:"color"`
}

func (m *TagModel) MapFromApi(api *sonatypeiq.ApiTagDTO) {
	m.ID = types.StringPointerValue(api.Id)
	m.Name = types.StringPointerValue(api.Name)
	m.Description = types.StringPointerValue(api.Description)
	m.Color = types.StringPointerValue(api.Color)
}

func (m *TagModel) MapToApi() *sonatypeiq.ApiTagDTO {
	api := sonatypeiq.NewApiTagDTOWithDefaults()
	api.Id = m.ID.ValueStringPointer()
	api.Name = m.Name.ValueStringPointer()
	api.Description = m.Description.ValueStringPointer()
	api.Color = m.Color.ValueStringPointer()
	return api
}

// ApplicationCategory
// ------------------------------------------------------------
type ApplicationCategory struct {
	TagModel
	OrganizationId types.String `tfsdk:"organization_id"`
}

// ApplicationCategories
// ------------------------------------------------------------
type ApplicationCategories struct {
	ID              types.String          `tfsdk:"id"`
	OrganiziationId types.String          `tfsdk:"organization_id"`
	Categories      []ApplicationCategory `tfsdk:"categories"`
}

func (m *ApplicationCategories) MapFromApi(api *[]sonatypeiq.ApiApplicationCategoryDTO) {
	m.Categories = make([]ApplicationCategory, 0)
	for _, category := range *api {
		m.Categories = append(m.Categories, ApplicationCategory{
			TagModel: TagModel{
				ID:          types.StringPointerValue(category.Id),
				Name:        types.StringPointerValue(category.Name),
				Description: types.StringPointerValue(category.Description),
				Color:       types.StringPointerValue(category.Color),
			},
			OrganizationId: types.StringPointerValue(category.OrganizationId),
		})
	}
}

// ApplicationCategoryModel
// ------------------------------------------------------------
type ApplicationCategoryModel struct {
	ApplicationCategory
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (m *ApplicationCategoryModel) MapFromApi(api *sonatypeiq.ApiApplicationCategoryDTO) {
	m.ID = types.StringPointerValue(api.Id)
	m.Name = types.StringPointerValue(api.Name)
	m.Description = types.StringPointerValue(api.Description)
	m.Color = types.StringPointerValue(api.Color)
	m.OrganizationId = types.StringPointerValue(api.OrganizationId)
}

func (m *ApplicationCategoryModel) MapToApi() *sonatypeiq.ApiApplicationCategoryDTO {
	api := sonatypeiq.NewApiApplicationCategoryDTOWithDefaults()
	api.Id = m.ID.ValueStringPointer()
	api.Name = m.Name.ValueStringPointer()
	api.Description = m.Description.ValueStringPointer()
	api.Color = m.Color.ValueStringPointer()
	api.OrganizationId = m.OrganizationId.ValueStringPointer()
	return api
}
