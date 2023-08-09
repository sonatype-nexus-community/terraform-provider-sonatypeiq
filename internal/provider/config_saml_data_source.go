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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &systemConfigDataSource{}
	_ datasource.DataSourceWithConfigure = &systemConfigDataSource{}
)

// ConfigSamlDataSource is a helper function to simplify the provider implementation.
func ConfigSamlDataSource() datasource.DataSource {
	return &configSamlDataSource{}
}

// configSamlDataSource is the data source implementation.
type configSamlDataSource struct {
	baseDataSource
}

type configSamlModel struct {
	ID           types.String `tfsdk:"id"`
	SamlMetadata types.String `tfsdk:"saml_metadata"`
}

// Metadata returns the data source type name.
func (d *configSamlDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_saml"
}

// Schema defines the schema for the data source.
func (d *configSamlDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get SAML Metadata for Sonatype IQ Server",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"saml_metadata": schema.StringAttribute{
				Description: "SAML Metadata for Sonatype IQ Server",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *configSamlDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data configSamlModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		d.auth,
	)

	// Make API Call
	saml_metadata, api_response, err := d.client.ConfigSAMLAPI.GetMetadata(ctx).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IQ System Configuration",
			err.Error(),
		)
		return
	}
	if api_response.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API Response", api_response.Status)
		return
	}

	data.ID = types.StringValue("placeholder")
	data.SamlMetadata = types.StringValue(saml_metadata)

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
