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

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &organizationDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationDataSource{}
)

// OrganizationDataSource is a helper function to simplify the provider implementation.
func OrganizationDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

// applicationsDataSource is the data source implementation.
type organizationDataSource struct {
	baseDataSource
}

// Metadata returns the data source type name.
func (d *organizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

// Schema defines the schema for the data source.
func (d *organizationDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch a specific Organization",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal ID of the Organization",
				Computed:    true,
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the Organization",
				Computed:    true,
			},
			"parent_organization_id": schema.StringAttribute{
				Description: "Internal ID of the Parent Organization if this Organization has a Parent Organization",
				Computed:    true,
			},
			"tags": schema.ListNestedAttribute{
				Description: "List of Tags belonging to this Organization",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"color": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data organizationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		d.auth,
	)

	if !data.ID.IsNull() {
		// Lookup By Org ID
		org, r, err := d.client.OrganizationsAPI.GetOrganization(ctx, data.ID.ValueString()).Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read IQ Organization by ID",
				err.Error(),
			)
			return
		}
		if r.StatusCode == 200 {
			var parentOrgId = types.StringNull()
			if org.ParentOrganizationId != nil {
				tflog.Debug(ctx, fmt.Sprintf("Parent Org Id is %s", *org.ParentOrganizationId))
				parentOrgId = types.StringValue(*org.ParentOrganizationId)
			}
			om := organizationModel{
				ID:                    types.StringValue(*org.Id),
				Name:                  types.StringValue(*org.Name),
				ParentOrganiziationId: parentOrgId,
			}
			for _, tag := range org.Tags {
				om.Tags = append(om.Tags, tagModel{
					ID:          types.StringValue(*tag.Id),
					Name:        types.StringValue(*tag.Name),
					Description: types.StringValue(*tag.Description),
					Color:       types.StringValue(*tag.Color),
				})
			}

			data = om
		}
	} else {
		resp.Diagnostics.AddError("No Organization ID provided ", "ID must be provided")
		return
	}

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}