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
	_ datasource.DataSource              = &applicationsDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationsDataSource{}
)

// ApplicationsDataSource is a helper function to simplify the provider implementation.
func ApplicationsDataSource() datasource.DataSource {
	return &applicationsDataSource{}
}

// applicationsDataSource is the data source implementation.
type applicationsDataSource struct {
	client *sonatypeiq.APIClient
	auth   sonatypeiq.BasicAuth
}

type applicationsDataSourceModel struct {
	Applications []applicationModel `tfsdk:"applications"`
}

type applicationModel struct {
	ID       types.String `tfsdk:"id"`
	PublicId types.String `tfsdk:"public_id"`
	Name     types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *applicationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_applications"
}

// Schema defines the schema for the data source.
func (d *applicationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"applications": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"public_id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *applicationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(SonatypeDataSourceData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Type",
			fmt.Sprintf("Expected provider.SonatypeDataSourceData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = config.client
	d.auth = config.auth
}

// Read refreshes the Terraform state with the latest data.
func (d *applicationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state applicationsDataSourceModel

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		d.auth,
	)

	applicationList, _, err := d.client.ApplicationsAPI.GetApplications(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IQ Applications",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Iterating %d Applications", len(applicationList.Applications)))

	for _, application := range applicationList.Applications {
		applicationState := applicationModel{
			ID:       types.StringValue(*application.Id),
			PublicId: types.StringValue(*application.PublicId),
			Name:     types.StringValue(*application.Name),
		}

		state.Applications = append(state.Applications, applicationState)

		tflog.Debug(ctx, fmt.Sprintf("   Appended: %p", state.Applications))
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
