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
	"net/url"
	"os"
	"strings"
	"terraform-provider-sonatypeiq/internal/provider/application"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/organization"
	"terraform-provider-sonatypeiq/internal/provider/role"
	"terraform-provider-sonatypeiq/internal/provider/scm"
	"terraform-provider-sonatypeiq/internal/provider/system"
	"terraform-provider-sonatypeiq/internal/provider/user"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

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
	Url         types.String `tfsdk:"url"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	VersionHint types.String `tfsdk:"version_hint"`
}

func (p *SonatypeIqProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sonatypeiq"
	resp.Version = p.version
}

func (p *SonatypeIqProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "Sonatype IQ Server URL, must start `http://` or `https://`, if not provided will attempt to fall back to environment variable `IQ_SERVER_URL`",
				Optional:            true,
				// Validators:          []validator.String{stringvalidator.LengthAtLeast(8)},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for Sonatype IQ Server, requires role/permissions scoped to the resources you wish to manage, if not provided will attempt to fall back to environment variable `IQ_SERVER_USERNAME`",
				Optional:            true,
				// Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for your user for Sonatype IQ Server, if not provided will attempt to fall back to environment variable `IQ_SERVER_PASSWORD`",
				Optional:            true,
				Sensitive:           true,
				// Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"version_hint": schema.StringAttribute{
				MarkdownDescription: `The version of Sonatype IQ Server you are connecting to, either in full as shown in the Server header (e.g. ` + "`1.205.4-01`" + `) or as just the release number (e.g. ` + "`205`" + `).

> [!NOTE]
> By default, this provider determines the version of Sonatype IQ Server automatically, from the _Server_ HTTP header. A reverse proxy or load balancer may strip or rewrite that header, in which case the version cannot be determined.

> [!TIP]
> Only set this if you receive a warning that the version of Sonatype IQ Server could not be determined, or if a reverse proxy rewrites the _Server_ header. When set, it takes precedence over what the server reports.`,
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(common.VersionHintExp, "must be the full version (e.g. 1.205.4-01) or the release number (e.g. 205)"),
				},
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

	iqUrl, username, password, versionHint := p.parseConfig(&config)

	p.validateConfig(resp, iqUrl, &config)
	if resp.Diagnostics.HasError() {
		return
	}

	ds := p.createClient(iqUrl, username, password)

	p.checkVersion(ctx, &ds, resp, versionHint)

	resp.DataSourceData = ds
	resp.ResourceData = ds
}

func (p *SonatypeIqProvider) parseConfig(config *SonatypeIqProviderModel) (string, string, string, *string) {
	iqUrl := os.Getenv("IQ_SERVER_URL")
	username := os.Getenv("IQ_SERVER_USERNAME")
	password := os.Getenv("IQ_SERVER_PASSWORD")

	if !config.Url.IsNull() && len(config.Url.ValueString()) > 0 {
		iqUrl = config.Url.ValueString()
	}

	if !config.Username.IsNull() && len(config.Username.ValueString()) > 0 {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() && len(config.Password.ValueString()) > 0 {
		password = config.Password.ValueString()
	}

	var versionHint *string
	if !config.VersionHint.IsNull() && len(config.VersionHint.ValueString()) > 0 {
		hint := config.VersionHint.ValueString()
		versionHint = &hint
	}

	return iqUrl, username, password, versionHint
}

func (p *SonatypeIqProvider) validateConfig(resp *provider.ConfigureResponse, iqUrl string, config *SonatypeIqProviderModel) {
	if len(iqUrl) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Unknown Sonatype IQ Server URL",
			"The provider is unable to work without a Sonatype IQ Server URL which should begin http:// or https://",
		)
	}

	if _, e := url.ParseRequestURI(iqUrl); e != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Invalid Sonatype IQ Server URL",
			"The provider is unable to work without a valid Sonatype IQ Server URL",
		)
	}

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
			"Password not supplied",
			"Administratrive credentials for your Sonatype IQ Server are required",
		)
	}
}

func (p *SonatypeIqProvider) createClient(iqUrl, username, password string) common.SonatypeDataSourceData {
	configuration := sonatypeiq.NewConfiguration()
	configuration.UserAgent = "sonatypeiq-terraform/" + p.version
	configuration.Servers = []sonatypeiq.ServerConfiguration{
		{
			URL:         iqUrl,
			Description: "Sonatype IQ Server",
		},
	}

	client := sonatypeiq.NewAPIClient(configuration)
	return common.SonatypeDataSourceData{
		Auth:    sonatypeiq.BasicAuth{UserName: username, Password: password},
		BaseUrl: strings.TrimRight(iqUrl, "/"),
		Client:  client,
	}
}

func (p *SonatypeIqProvider) checkVersion(ctx context.Context, ds *common.SonatypeDataSourceData, resp *provider.ConfigureResponse, versionHint *string) {
	ds.CheckConnectionAndGetVersion(ctx, &resp.Diagnostics, versionHint)
	tflog.Info(ctx, fmt.Sprintf("Detected Sonatype IQ Server to be version %v", ds.IqVersion))

	if resp.Diagnostics.HasError() {
		return
	}

	if ds.IqVersion == 0 {
		resp.Diagnostics.AddWarning(
			`Could not determine the version of Sonatype IQ Server`,
			`The _Server_ HTTP header that identifies the version was missing or not recognised, for example because a reverse proxy or load balancer rewrites it. `+
				`Everything will work, but features that depend on the version of Sonatype IQ Server are not checked. `+
				`Set 'version_hint' in the provider configuration to the version you are running (e.g. '1.205.4-01' or '205') to turn those checks back on.`,
		)
		return
	}

	if ds.IqVersion < 186 {
		resp.Diagnostics.AddWarning(
			`You are running against Sonatype IQ Server version older than 186`,
			`This provide has not been validated against versions older than 186 - things will probably work fine, but proceed with caution.`,
		)
	}
}

func (p *SonatypeIqProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		application.NewApplicationResource,
		application.NewApplicationRoleMembershipResource,
		organization.NewApplicationCategoryResource,
		organization.NewOrganizationResource,
		organization.NewOrganizationRoleMembershipResource,
		role.NewGlobalRoleMembershipResource,
		role.NewRoleResource,
		scm.NewSourceControlResource,
		system.NewConfigCrowdResource,
		system.NewConfigMailResource,
		system.NewConfigProductLicenseResource,
		system.NewConfigProxyServerResource,
		system.NewConfigSamlResource,
		system.NewSystemConfigResource,
		user.NewUserResource,
		user.NewUserTokenResource,
	}
}

func (p *SonatypeIqProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		application.ApplicationDataSource,
		application.ApplicationsDataSource,
		application.ApplicationCategoriesDataSource,
		organization.OrganizationDataSource,
		organization.OrganizationsDataSource,
		role.RoleDataSource,
		system.ConfigSamlDataSource,
		system.SystemConfigDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SonatypeIqProvider{
			version: version,
		}
	}
}
