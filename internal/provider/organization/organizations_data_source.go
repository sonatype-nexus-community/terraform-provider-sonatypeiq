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

package organization

import (
	"context"
	"fmt"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &organizationsDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationsDataSource{}
)

// OrganizationsDataSource is a helper function to simplify the provider implementation.
func OrganizationsDataSource() datasource.DataSource {
	return &organizationsDataSource{}
}

// applicationsDataSource is the data source implementation.
type organizationsDataSource struct {
	common.BaseDataSource
}

type organizationsDataSourceModel struct {
	ID            types.String              `tfsdk:"id"`
	Organizations []model.OrganizationModel `tfsdk:"organizations"`
}

// type organizationModel struct {
// 	ID                    types.String `tfsdk:"id"`
// 	Name                  types.String `tfsdk:"name"`
// 	ParentOrganiziationId types.String `tfsdk:"parent_organization_id"`
// 	Tags                  []tagModel   `tfsdk:"tags"`
// }

// type tagModel struct {
// 	ID          types.String `tfsdk:"id"`
// 	Name        types.String `tfsdk:"name"`
// 	Description types.String `tfsdk:"description"`
// 	Color       types.String `tfsdk:"color"`
// }

// Metadata returns the data source type name.
func (d *organizationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizations"
}

// Schema defines the schema for the data source.
func (d *organizationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get all Organizations",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"organizations": schema.ListNestedAttribute{
				Description: "List of Organizations",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Internal ID of the Organization",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the Organization",
							Computed:    true,
						},
						"parent_organization_id": schema.StringAttribute{
							Description: "Internal ID of the Organization to which this Organization belongs",
							Computed:    true,
						},
						"tags": schema.ListNestedAttribute{
							Description: "List of any Tags associated to this Organization",
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
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *organizationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state organizationsDataSourceModel

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		d.Auth,
	)

	orgList, _, err := d.Client.OrganizationsAPI.GetOrganizations(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IQ Organizations",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Iterating %d Organizations", len(orgList.Organizations)))

	for _, organization := range orgList.Organizations {
		var parentOrgId = types.StringNull()
		if organization.ParentOrganizationId != nil {
			tflog.Debug(ctx, fmt.Sprintf("Parent Org Id is %s", *organization.ParentOrganizationId))
			parentOrgId = types.StringValue(*organization.ParentOrganizationId)
		}
		organizationState := model.OrganizationModel{
			ID:                    types.StringValue(*organization.Id),
			Name:                  types.StringValue(*organization.Name),
			ParentOrganiziationId: parentOrgId,
		}

		for _, tag := range organization.Tags {
			organizationState.Tags = append(organizationState.Tags, model.TagModel{
				ID:          types.StringValue(*tag.Id),
				Name:        types.StringValue(*tag.Name),
				Description: types.StringValue(*tag.Description),
				Color:       types.StringValue(*tag.Color),
			})
		}

		state.Organizations = append(state.Organizations, organizationState)

		tflog.Debug(ctx, fmt.Sprintf("   Appended: %p", state.Organizations))
	}

	// For test framework
	state.ID = types.StringValue("placeholder")

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
