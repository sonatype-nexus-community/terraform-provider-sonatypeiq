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

// SystemConfigDataSource is a helper function to simplify the provider implementation.
func SystemConfigDataSource() datasource.DataSource {
	return &systemConfigDataSource{}
}

// systemConfigDataSource is the data source implementation.
type systemConfigDataSource struct {
	baseDataSource
}

type systemConfigModel struct {
	ID           types.String `tfsdk:"id"`
	BaseURL      types.String `tfsdk:"base_url"`
	ForceBaseURL types.Bool   `tfsdk:"force_base_url"`
}

// Metadata returns the data source type name.
func (d *systemConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_config"
}

// Schema defines the schema for the data source.
func (d *systemConfigDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get System Configuration",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"base_url": schema.StringAttribute{
				Description: "Base URL for Sonatype IQ Server",
				Computed:    true,
				Optional:    true,
			},
			"force_base_url": schema.BoolAttribute{
				Description: "Should the Base URL be forced?",
				Computed:    true,
				Optional:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *systemConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data systemConfigModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		d.auth,
	)

	// Lookup System Configuration
	config_request := d.client.ConfigurationAPI.GetConfiguration(ctx)
	config_request = config_request.Property([]sonatypeiq.SystemConfigProperty{
		"baseUrl", "forceBaseUrl",
	})
	config, r, err := config_request.Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IQ System Configuration",
			err.Error(),
		)
		return
	}
	if r.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API Response", r.Status)
		return
	}

	if config.BaseUrl.IsSet() {
		data.BaseURL = types.StringValue(config.GetBaseUrl())
	}
	if config.ForceBaseUrl.IsSet() {
		data.ForceBaseURL = types.BoolValue(config.GetForceBaseUrl())
	}

	data.ID = types.StringValue("placeholder")

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
