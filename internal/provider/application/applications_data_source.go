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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	"github.com/sonatype-nexus-community/terraform-provider-shared/schema"
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
	common.BaseDataSource
}

// Metadata returns the data source type name.
func (d *applicationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_applications"
}

// Schema defines the schema for the data source.
func (d *applicationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Description: "Use this data source to get all Applications",
		Attributes: map[string]tfschema.Attribute{
			"id": schema.DataSourceComputedString("The ID of this resource."),
			"applications": schema.DataSourceComputedListNestedAttribute(
				"List of Applications",
				tfschema.NestedAttributeObject{
					Attributes: map[string]tfschema.Attribute{
						"id":                schema.DataSourceComputedString("Internal ID of the Application"),
						"public_id":         schema.DataSourceComputedString("Public ID of the Application"),
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
				},
			),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *applicationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.ApplicationsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := d.Client.ApplicationsAPI.GetApplications(d.AuthContext(ctx)).Execute()

	if err != nil {
		errors.HandleAPIError(
			common.ERR_FAILED_READING_APPLICATIONS,
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.AddAPIErrorDiagnostic(&resp.Diagnostics, "read", "Applications", httpResponse, err)
		return
	}

	// Assign Response Data to State
	data.ID = types.StringValue("all-applications")
	data.MapFromApi(apiResponse)

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
