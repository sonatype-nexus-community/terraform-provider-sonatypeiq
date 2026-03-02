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

package application

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
	sharedutil "github.com/sonatype-nexus-community/terraform-provider-shared/util"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &applicationDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationDataSource{}
)

// ApplicationDataSource is a helper function to simplify the provider implementation.
func ApplicationDataSource() datasource.DataSource {
	return &applicationDataSource{}
}

// applicationsDataSource is the data source implementation.
type applicationDataSource struct {
	common.BaseDataSource
}

// Metadata returns the data source type name.
func (d *applicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

// Schema defines the schema for the data source.
func (d *applicationDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get an Application",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal ID of the Application",
				Computed:    true,
				Optional:    true,
			},
			"public_id": schema.StringAttribute{
				Description: "Public ID of the Application",
				Computed:    true,
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the Application",
				Computed:    true,
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Internal ID of the Organization to which this Application belongs",
				Computed:    true,
			},
			"contact_user_name": schema.StringAttribute{
				Description: "User Name of the Contact for the Application",
				Computed:    true,
				Optional:    true,
			},
			"application_tags": sharedrschema.DataSourceComputedListNestedAttribute(
				"List of Tags applied to this Application",
				schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Internal ID of the Application-Tag link",
							Computed:    true,
						},
						"tag_id": schema.StringAttribute{
							Description: "Internal ID of the Tag",
							Computed:    true,
						},
						"application_id": schema.StringAttribute{
							Description: "Internal ID of the Application",
							Computed:    true,
						},
					},
				},
			),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *applicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.ApplicationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		d.Auth,
	)

	var app *sonatypeiq.ApiApplicationDTO
	var r *http.Response
	var err error

	if !data.ID.IsNull() {
		// Lookup By Application ID
		app, r, err = d.Client.ApplicationsAPI.GetApplication(ctx, data.ID.ValueString()).Execute()

		if err != nil {
			sharederr.HandleAPIError("Unable to read IQ Application by ID", &err, r, &resp.Diagnostics)
			return
		}
		if r.StatusCode != http.StatusOK {
			sharederr.AddAPIErrorDiagnostic(&resp.Diagnostics, "read", "Application", r, err)
			return
		}

	} else if !data.PublicId.IsNull() {
		// Lookup By Application Public ID
		var apps *sonatypeiq.ApiApplicationListDTO
		get_apps_req := d.Client.ApplicationsAPI.GetApplications(ctx)
		get_apps_req = get_apps_req.PublicId([]string{data.PublicId.ValueString()})
		apps, r, err = get_apps_req.Execute()
		if err != nil {
			sharederr.HandleAPIError("Unable to read IQ Application by Public ID", &err, r, &resp.Diagnostics)
			return
		}
		if r.StatusCode != http.StatusOK {
			sharederr.AddAPIErrorDiagnostic(&resp.Diagnostics, "read", "Application", r, err)
			return
		}
		if len(apps.Applications) == 1 {
			app = &apps.Applications[0]
		} else if len(apps.Applications) > 1 {
			sharederr.AddValidationDiagnostic(&resp.Diagnostics, "Application Public ID", "More than one Application matched the provided Public ID")
			return
		}
	} else {
		sharederr.AddValidationDiagnostic(&resp.Diagnostics, "Application Lookup", "ID or Public ID must be provided")
		return
	}

	if app == nil {
		sharederr.AddNotFoundDiagnostic(&resp.Diagnostics, "Application", data.ID.ValueString())
		return
	}

	appModel := model.ApplicationModel{
		ID:              sharedutil.StringPtrToValue(app.Id),
		PublicId:        sharedutil.StringPtrToValue(app.PublicId),
		Name:            sharedutil.StringPtrToValue(app.Name),
		OrganizationId:  sharedutil.StringPtrToValue(app.OrganizationId),
		ContactUserName: sharedutil.StringPtrToValue(app.ContactUserName),
	}
	for _, tag := range app.ApplicationTags {
		appModel.ApplicationTags = append(appModel.ApplicationTags, model.ApplicationTagLinkModel{
			ID:            sharedutil.StringPtrToValue(tag.Id),
			TagId:         sharedutil.StringPtrToValue(tag.TagId),
			ApplicationId: sharedutil.StringPtrToValue(tag.ApplicationId),
		})
	}

	// Set state
	diags := resp.State.Set(ctx, &appModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
