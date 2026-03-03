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
	"fmt"
	"net/http"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"

	// sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	"github.com/sonatype-nexus-community/terraform-provider-shared/schema"
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
	resp.Schema = tfschema.Schema{
		Description: "Use this data source to get an Application",
		Attributes: map[string]tfschema.Attribute{
			"id":                schema.DataSourceOptionalString("Internal ID of the Application"),
			"public_id":         schema.DataSourceOptionalString("Public ID of the Application"),
			"name":              schema.DataSourceComputedString("Name of the Application"),
			"organization_id":   schema.DataSourceComputedString("Internal ID of the Organization to which this Application belongs"),
			"contact_user_name": schema.DataSourceComputedString("User Name of the Contact for the Application"),
			"application_tags": schema.DataSourceComputedListNestedAttribute(
				"List of Tags applied to this Application",
				tfschema.NestedAttributeObject{
					Attributes: map[string]tfschema.Attribute{
						"id":             schema.DataSourceComputedString("Internal ID of the Application-Tag link"),
						"tag_id":         schema.DataSourceComputedString("Internal ID of the Tag"),
						"application_id": schema.DataSourceComputedString("Internal ID of the Application"),
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
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Lookup
	var foundApplication *sonatypeiq.ApiApplicationDTO
	var httpResponse *http.Response
	var err error
	if !data.ID.IsNull() {
		foundApplication, httpResponse, err = d.Client.ApplicationsAPI.GetApplication(d.AuthContext(ctx), data.ID.ValueString()).Execute()

		if err != nil || httpResponse.StatusCode != http.StatusOK {
			errors.HandleAPIError("Unable to read IQ Application by ID", &err, httpResponse, &resp.Diagnostics)
			return
		}
	} else if !data.PublicId.IsNull() {
		var apiResponse *sonatypeiq.ApiApplicationListDTO
		apiResponse, httpResponse, err = d.Client.ApplicationsAPI.GetApplications(d.AuthContext(ctx)).PublicId([]string{data.PublicId.ValueString()}).Execute()

		if err != nil || httpResponse.StatusCode != http.StatusOK {
			errors.HandleAPIError("Unable to read IQ Applications to find by Public ID", &err, httpResponse, &resp.Diagnostics)
			return
		} else if len(apiResponse.Applications) != 1 {
			errors.HandleAPIWarning("No unique Application found with supplied Public ID", nil, httpResponse, &resp.Diagnostics)
			return
		}

		foundApplication = &apiResponse.Applications[0]
	} else {
		errors.AddValidationDiagnostic(&resp.Diagnostics, "Application Lookup", "ID or Public ID must be provided")
		return
	}

	// Map api response to State
	data.MapFromApi(foundApplication)

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
