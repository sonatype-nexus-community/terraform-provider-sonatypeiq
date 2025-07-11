/**
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

package application

import (
	"context"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &applicationCategoriesDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationCategoriesDataSource{}
)

// ApplicationCategoriesDataSource is a helper function to simplify the provider implementation.
func ApplicationCategoriesDataSource() datasource.DataSource {
	return &applicationCategoriesDataSource{}
}

// applicationsDataSource is the data source implementation.
type applicationCategoriesDataSource struct {
	common.BaseDataSource
}

type applicationCategoriesModel struct {
	ID              types.String     `tfsdk:"id"`
	OrganiziationId types.String     `tfsdk:"organization_id"`
	Categories      []model.TagModel `tfsdk:"categories"`
}

// Metadata returns the data source type name.
func (d *applicationCategoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_categories"
}

// Schema defines the schema for the data source.
func (d *applicationCategoriesDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get Application Categories for an Organization",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Internal ID of the Organization to which this Application belongs - use 'ROOT_ORGANIZATION_ID' for the Root Organization",
				Required:    true,
			},
			"categories": schema.ListNestedAttribute{
				Description: "List of Categories defined for this Organization",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Internal ID of the Tag",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the Tag",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the Tag",
							Computed:    true,
						},
						"color": schema.StringAttribute{
							Description: "Color of the Tag",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *applicationCategoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data applicationCategoriesModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		d.Auth,
	)

	categories, api_response, err := d.Client.ApplicationCategoriesAPI.GetTags(ctx, data.OrganiziationId.ValueString()).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IQ Application Categories for Organization",
			err.Error(),
		)
		return
	}
	if api_response.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API Response", api_response.Status)
		return
	}

	for _, category := range categories {
		data.Categories = append(data.Categories, model.TagModel{
			ID:          types.StringValue(*category.Id),
			Name:        types.StringValue(*category.Name),
			Description: types.StringValue(*category.Description),
			Color:       types.StringValue(*category.Color),
		})
	}

	// For test framework
	data.ID = types.StringValue("placeholder")

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
