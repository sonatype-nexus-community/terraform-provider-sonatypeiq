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

// OrganizationsModel
// ------------------------------------------------------------
type OrganizationsModel struct {
	ID            types.String        `tfsdk:"id"`
	Organizations []OrganizationModel `tfsdk:"organizations"`
}

func (m *OrganizationsModel) MapFromApi(api *sonatypeiq.ApiOrganizationListDTO) {
	m.Organizations = make([]OrganizationModel, 0)
	for _, apiOrg := range api.Organizations {
		org := OrganizationModel{}
		org.MapFromApi(&apiOrg)
		m.Organizations = append(m.Organizations, org)
	}
}

// OrganizationModel
// ------------------------------------------------------------
type OrganizationModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	ParentOrganiziationId types.String `tfsdk:"parent_organization_id"`
	Categories            []TagModel   `tfsdk:"categories"`
}

func (m *OrganizationModel) MapFromApi(api *sonatypeiq.ApiOrganizationDTO) {
	m.ID = types.StringPointerValue(api.Id)
	m.Name = types.StringPointerValue(api.Name)
	m.ParentOrganiziationId = types.StringPointerValue(api.ParentOrganizationId)
	m.Categories = make([]TagModel, 0)
	for _, apiTag := range api.Tags {
		t := TagModel{}
		t.MapFromApi(&apiTag)
		m.Categories = append(m.Categories, t)
	}
}

// OrganizationModelResource
// ------------------------------------------------------------
type OrganizationModelResource struct {
	OrganizationModel
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (m *OrganizationModelResource) MapToApi() *sonatypeiq.ApiOrganizationDTO {
	api := sonatypeiq.NewApiOrganizationDTOWithDefaults()
	api.Id = m.ID.ValueStringPointer()
	api.Name = m.Name.ValueStringPointer()
	api.ParentOrganizationId = m.ParentOrganiziationId.ValueStringPointer()
	// api.Tags = make([]sonatypeiq.ApiTagDTO, 0)
	// for _, mCat := range m.Categories {
	// 	tag := mCat.MapToApi()
	// 	api.Tags = append(api.Tags, *tag)
	// }
	return api
}
