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

import "github.com/hashicorp/terraform-plugin-framework/types"

type ApplicationModel struct {
	ID              types.String              `tfsdk:"id"`
	PublicId        types.String              `tfsdk:"public_id"`
	Name            types.String              `tfsdk:"name"`
	OrganizationId  types.String              `tfsdk:"organization_id"`
	ContactUserName types.String              `tfsdk:"contact_user_name"`
	ApplicationTags []ApplicationTagLinkModel `tfsdk:"application_tags"`
}

type ApplicationModellResource struct {
	ID              types.String `tfsdk:"id"`
	PublicId        types.String `tfsdk:"public_id"`
	Name            types.String `tfsdk:"name"`
	OrganizationId  types.String `tfsdk:"organization_id"`
	ContactUserName types.String `tfsdk:"contact_user_name"`
	LastUpdated     types.String `tfsdk:"last_updated"`
}

type ApplicationTagLinkModel struct {
	ID            types.String `tfsdk:"id"`
	TagId         types.String `tfsdk:"tag_id"`
	ApplicationId types.String `tfsdk:"application_id"`
}
