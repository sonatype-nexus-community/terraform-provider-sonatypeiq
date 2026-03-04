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

package organization

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	"github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &organizationsDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationsDataSource{}
)

// OrganizationsDataSource is a helper function to simplify the provider implementation.
func OrganizationsDataSource() datasource.DataSource {
	return &organizationsDataSource{}
}

// applicationsDataSource is the data source implementation.
type organizationsDataSource struct {
	common.BaseDataSource
}

// Metadata returns the data source type name.
func (d *organizationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizations"
}

// Schema defines the schema for the data source.
func (d *organizationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Description: "Use this data source to get all Organizations",
		Attributes: map[string]tfschema.Attribute{
			"id": schema.DataSourceComputedString("The ID of this resource."),
			"organizations": schema.DataSourceComputedListNestedAttribute(
				"List of Organizations",
				tfschema.NestedAttributeObject{
					Attributes: map[string]tfschema.Attribute{
						"id":                     schema.DataSourceComputedString("Internal ID of the Organization"),
						"name":                   schema.DataSourceComputedString("Name of the Organization"),
						"parent_organization_id": schema.DataSourceComputedString("Internal ID of the Organization to which this Organization belongs"),
						"categories": schema.DataSourceComputedListNestedAttribute(
							"List of Application Categories defined in this Organization",
							tfschema.NestedAttributeObject{
								Attributes: map[string]tfschema.Attribute{
									"id":          schema.DataSourceComputedString("Internal ID of the Tag"),
									"name":        schema.DataSourceComputedString("Name of the Tag"),
									"description": schema.DataSourceComputedString("Description of the Tag"),
									"color":       schema.DataSourceComputedString("Color of the Tag"),
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
func (d *organizationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.OrganizationsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := d.Client.OrganizationsAPI.GetOrganizations(d.AuthContext(ctx)).Execute()

	if err != nil {
		errors.HandleAPIError(
			common.ERR_FAILED_READING_ORGANIZATIONS,
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.AddAPIErrorDiagnostic(&resp.Diagnostics, "read", "Organizations", httpResponse, err)
		return
	}

	// Assign Response Data to State
	data.ID = types.StringValue("all-organizations")
	data.MapFromApi(apiResponse)

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
