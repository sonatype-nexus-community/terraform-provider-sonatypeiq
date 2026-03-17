/**
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
	_ datasource.DataSource              = &applicationCategoriesDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationCategoriesDataSource{}
)

// ApplicationCategoriesDataSource is a helper function to simplify the provider implementation.
func ApplicationCategoriesDataSource() datasource.DataSource {
	return &applicationCategoriesDataSource{}
}

// applicationsDataSource is the data source implementation.
type applicationCategoriesDataSource struct {
	common.BaseDataSource
}

// Metadata returns the data source type name.
func (d *applicationCategoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_categories"
}

// Schema defines the schema for the data source.
func (d *applicationCategoriesDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Description: "Use this data source to get Application Categories for an Organization",
		Attributes: map[string]tfschema.Attribute{
			"id":              schema.DataSourceComputedString("The ID of this resource."),
			"organization_id": schema.DataSourceRequiredString("Internal ID of the Organization to which this Application belongs - use `ROOT_ORGANIZATION_ID` for the Root Organization"),
			"categories": schema.DataSourceComputedListNestedAttribute(
				"List of Categories defined for this Organization",
				tfschema.NestedAttributeObject{
					Attributes: map[string]tfschema.Attribute{
						"id":              schema.DataSourceComputedString("Internal ID of the Application Category"),
						"name":            schema.DataSourceComputedString("Name of the Application Category"),
						"description":     schema.DataSourceComputedString("Description of the Application Category"),
						"organization_id": schema.DataSourceComputedString("Organization ID this Application Category belongs to"),
						"color":           schema.DataSourceComputedString("Color of the Application Category"),
					},
				},
			),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *applicationCategoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.ApplicationCategories
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := d.Client.ApplicationCategoriesAPI.GetTags(d.AuthContext(ctx), data.OrganiziationId.ValueString()).Execute()

	if err != nil {
		errors.HandleAPIError(
			common.ERR_FAILED_READING_APPLICATION_CATEGORIES_FOR_ORG,
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.AddAPIErrorDiagnostic(&resp.Diagnostics, "read", "Application Categories", httpResponse, err)
		return
	}

	// Assign Response Data to State
	data.ID = types.StringValue(fmt.Sprintf("application-categories-%s", data.OrganiziationId.ValueString()))
	data.MapFromApi(&apiResponse)

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
