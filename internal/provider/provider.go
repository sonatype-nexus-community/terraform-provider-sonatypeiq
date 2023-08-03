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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// Ensure SonatypeIqProvider satisfies various provider interfaces.
var _ provider.Provider = &SonatypeIqProvider{}

// SonatypeIqProvider defines the provider implementation.
type SonatypeIqProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// SonatypeIqProviderModel describes the provider data model.
type SonatypeIqProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *SonatypeIqProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sonatype"
	resp.Version = p.version
}

func (p *SonatypeIqProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Sonatype IQ Server URL",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Administrator Username for Sonatype IQ Server",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for your Administrator user for Sonatype IQ Server",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *SonatypeIqProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config SonatypeIqProviderModel

	diags := req.Config.Get(ctx, &config)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate Provider Configuration
	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Sonatype IQ Server Host",
			"The provider is unable to work without a Sonatype IQ Server URL which should begin http:// or https://",
		)
	}

	// if _, error := url.ParseRequestURI(config.Host.ValueString()); error != nil {
	// 	resp.Diagnostics.AddAttributeError(
	// 		path.Root("host"),
	// 		"Invalid Sonatype IQ Server Host",
	// 		"The provider is unable to work without a valid Sonatype IQ Server URL",
	// 	)
	// }

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Username not supplied",
			"Administratrive credentials for your Sonatype IQ Server are required",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Username not supplied",
			"Administratrive credentials for your Sonatype IQ Server are required",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Example client configuration for data sources and resources
	configuration := sonatypeiq.NewConfiguration()
	configuration.Host = config.Host.ValueString()
	configuration.UserAgent = "sonatype-terraform-pf/" + p.version
	configuration.Servers = []sonatypeiq.ServerConfiguration{
		{
			URL:         "https://" + config.Host.ValueString(),
			Description: "Default Sonatype IQ Server",
		},
	}

	client := sonatypeiq.NewAPIClient(configuration)
	resp.DataSourceData = SonatypeDataSourceData{
		client: client,
		auth:   sonatypeiq.BasicAuth{UserName: config.Username.ValueString(), Password: config.Password.ValueString()},
	}
	resp.ResourceData = SonatypeDataSourceData{
		client: client,
		auth:   sonatypeiq.BasicAuth{UserName: config.Username.ValueString(), Password: config.Password.ValueString()},
	}
}

func (p *SonatypeIqProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewApplicationResource,
	}
}

func (p *SonatypeIqProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		ApplicationsDataSource,
		OrganizationDataSource,
		OrganizationsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SonatypeIqProvider{
			version: version,
		}
	}
}
