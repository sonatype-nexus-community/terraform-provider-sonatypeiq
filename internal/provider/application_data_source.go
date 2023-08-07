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
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
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
	baseDataSource
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
			"application_tags": schema.ListNestedAttribute{
				Description: "List of Tags applied to this Application",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
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
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *applicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data applicationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		d.auth,
	)

	var app *sonatypeiq.ApiApplicationDTO
	var r *http.Response
	var err error

	if !data.ID.IsNull() {
		// Lookup By Application ID
		app, r, err = d.client.ApplicationsAPI.GetApplication(ctx, data.ID.ValueString()).Execute()

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read IQ Application by ID",
				err.Error(),
			)
			return
		}
		if r.StatusCode != 200 {
			resp.Diagnostics.AddError("Unexpected API Response", r.Status)
			return
		}

	} else if !data.PublicId.IsNull() {
		// Lookup By Application Public ID
		var apps *sonatypeiq.ApiApplicationListDTO
		get_apps_req := d.client.ApplicationsAPI.GetApplications(ctx)
		get_apps_req = get_apps_req.PublicId([]string{data.PublicId.ValueString()})
		apps, r, err = get_apps_req.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read IQ Application by Public ID",
				err.Error(),
			)
			return
		}
		if r.StatusCode != 200 {
			resp.Diagnostics.AddError("Unexpected API Response", r.Status)
			return
		}
		if len(apps.Applications) == 1 {
			app = &apps.Applications[0]
		} else if len(apps.Applications) > 1 {
			resp.Diagnostics.AddError("More than one Applications matched the Public ID", r.Status)
			return
		}
	} else {
		resp.Diagnostics.AddError("No Application ID or Public ID provided ", "ID or Public ID must be provided")
		return
	}

	if app == nil {
		resp.Diagnostics.AddError("No Application found", "No Application found with the provided ID or Public ID")
		return
	}

	var contactUserName = types.StringNull()
	if app.ContactUserName != nil {
		contactUserName = types.StringValue(*app.ContactUserName)
	}
	appModel := applicationModel{
		ID:              types.StringValue(*app.Id),
		PublicId:        types.StringValue(*app.PublicId),
		Name:            types.StringValue(*app.Name),
		OrganizationId:  types.StringValue(*app.OrganizationId),
		ContactUserName: contactUserName,
	}
	for _, tag := range app.ApplicationTags {
		appModel.ApplicationTags = append(appModel.ApplicationTags, applicationTagLinkModel{
			ID:            types.StringValue(*tag.Id),
			TagId:         types.StringValue(*tag.TagId),
			ApplicationId: types.StringValue(*tag.ApplicationId),
		})
	}

	// Set state
	diags := resp.State.Set(ctx, &appModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
