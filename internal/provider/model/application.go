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

// ApplicationModel
// ------------------------------------------------------------
type ApplicationModel struct {
	ID              types.String              `tfsdk:"id"`
	PublicId        types.String              `tfsdk:"public_id"`
	Name            types.String              `tfsdk:"name"`
	OrganizationId  types.String              `tfsdk:"organization_id"`
	ContactUserName types.String              `tfsdk:"contact_user_name"`
	ApplicationTags []ApplicationTagLinkModel `tfsdk:"application_tags"`
}

func (m *ApplicationModel) MapFromApi(api *sonatypeiq.ApiApplicationDTO) {
	m.ID = types.StringPointerValue(api.Id)
	m.PublicId = types.StringPointerValue(api.PublicId)
	m.Name = types.StringPointerValue(api.Name)
	m.OrganizationId = types.StringPointerValue(api.OrganizationId)
	m.ContactUserName = types.StringPointerValue(api.ContactUserName)
	m.ApplicationTags = make([]ApplicationTagLinkModel, 0)
	for _, tagLink := range api.ApplicationTags {
		tl := &ApplicationTagLinkModel{}
		tl.MapFromApi(&tagLink)
		m.ApplicationTags = append(m.ApplicationTags, *tl)
	}
}

// ApplicationsModel
// ------------------------------------------------------------
type ApplicationsModel struct {
	ID           types.String       `tfsdk:"id"`
	Applications []ApplicationModel `tfsdk:"applications"`
}

func (m *ApplicationsModel) MapFromApi(api *sonatypeiq.ApiApplicationListDTO) {
	m.Applications = make([]ApplicationModel, 0)
	for _, apiApp := range api.Applications {
		app := ApplicationModel{}
		app.MapFromApi(&apiApp)
		m.Applications = append(m.Applications, app)
	}
}

// ApplicationModelResource
// ------------------------------------------------------------
type ApplicationModelResource struct {
	ID              types.String `tfsdk:"id"`
	PublicId        types.String `tfsdk:"public_id"`
	Name            types.String `tfsdk:"name"`
	OrganizationId  types.String `tfsdk:"organization_id"`
	ContactUserName types.String `tfsdk:"contact_user_name"`
	LastUpdated     types.String `tfsdk:"last_updated"`
}

// ApplicationTagLinkModel
// ------------------------------------------------------------
type ApplicationTagLinkModel struct {
	ID            types.String `tfsdk:"id"`
	TagId         types.String `tfsdk:"tag_id"`
	ApplicationId types.String `tfsdk:"application_id"`
}

func (m *ApplicationTagLinkModel) MapFromApi(api *sonatypeiq.ApiApplicationTagDTO) {
	m.ID = types.StringPointerValue(api.Id)
	m.TagId = types.StringPointerValue(api.TagId)
	m.ApplicationId = types.StringPointerValue(api.ApplicationId)
}
