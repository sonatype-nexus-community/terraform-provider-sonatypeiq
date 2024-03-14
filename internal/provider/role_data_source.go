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

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &roleDataSource{}
	_ datasource.DataSourceWithConfigure = &roleDataSource{}
)

// RoleDataSource is a helper function to simplify the provider implementation.
func RoleDataSource() datasource.DataSource {
	return &roleDataSource{}
}

// roleDataSource is the data source implementation.
type roleDataSource struct {
	baseDataSource
}

type roleDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Schema defines the schema for the data source.
func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get a role",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Description: "The role name",
				Required:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data roleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		d.auth,
	)

	roleList, _, err := d.client.RolesAPI.GetRoles(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IQ Roles",
			err.Error(),
		)
		return
	}

	for _, role := range roleList.Roles {
		if role.GetName() == data.Name.ValueString() {
			data.ID = types.StringValue(role.GetId())
		}
	}

	if data.ID.IsNull() {
		resp.Diagnostics.AddError(
			"Unable to find role",
			fmt.Sprintf("Role '%s' does not exist", data.Name.ValueString()),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
