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

type ApplicationCategory struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Color          types.String `tfsdk:"color"`
}

type ApplicationCategoryModel struct {
	ApplicationCategory
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (m *ApplicationCategoryModel) MapFromApi(api *sonatypeiq.ApiApplicationCategoryDTO) {
	m.ID = sharedutil.StringPtrToValue(api.Id)
	m.Name = sharedutil.StringPtrToValue(api.Name)
	m.Description = sharedutil.StringPtrToValue(api.Description)
	m.Color = sharedutil.StringPtrToValue(api.Color)
	m.OrganizationId = sharedutil.StringPtrToValue(api.OrganizationId)
}

func (m *ApplicationCategoryModel) MapToApi(api *sonatypeiq.ApiApplicationCategoryDTO) {
	api.Name = sharedutil.StringToPtr(m.Name.ValueString())
	api.Description = sharedutil.StringToPtr(m.Description.ValueString())
	api.Color = sharedutil.StringToPtr(m.Color.ValueString())
}
