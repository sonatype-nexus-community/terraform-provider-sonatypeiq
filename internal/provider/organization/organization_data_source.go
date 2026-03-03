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
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	"github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &organizationDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationDataSource{}
)

// OrganizationDataSource is a helper function to simplify the provider implementation.
func OrganizationDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

// organizationsDataSource is the data source implementation.
type organizationDataSource struct {
	common.BaseDataSource
}

// Metadata returns the data source type name.
func (d *organizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

// Schema defines the schema for the data source.
func (d *organizationDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Description: "Use this data source to get an Organization",
		Attributes: map[string]tfschema.Attribute{
			"id":                     schema.DataSourceOptionalString("Internal ID of the Organization"),
			"name":                   schema.DataSourceOptionalString("Name of the Organization"),
			"parent_organization_id": schema.DataSourceComputedString("Internal ID of the Parent Organization if this Organization has a Parent Organization"),
			"tags": schema.DataSourceComputedListNestedAttribute(
				"List of Tags associated to this Organization",
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
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.OrganizationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Lookup
	var foundOrganization *sonatypeiq.ApiOrganizationDTO
	var httpResponse *http.Response
	var err error

	if !data.ID.IsNull() {
		foundOrganization, httpResponse, err = d.Client.OrganizationsAPI.GetOrganization(d.AuthContext(ctx), data.ID.ValueString()).Execute()

		if err != nil || httpResponse.StatusCode != http.StatusOK {
			errors.HandleAPIError("Unable to read IQ Organization by ID", &err, httpResponse, &resp.Diagnostics)
			return
		}
	} else if !data.Name.IsNull() {
		var apiResponse *sonatypeiq.ApiOrganizationListDTO
		apiResponse, httpResponse, err = d.Client.OrganizationsAPI.GetOrganizations(d.AuthContext(ctx)).OrganizationName([]string{data.Name.ValueString()}).Execute()

		if err != nil || httpResponse.StatusCode != http.StatusOK {
			errors.HandleAPIError("Unable to read IQ Organizations to find by Name", &err, httpResponse, &resp.Diagnostics)
			return
		} else if len(apiResponse.Organizations) != 1 {
			errors.HandleAPIWarning("No unique Organization found with supplied Name", nil, httpResponse, &resp.Diagnostics)
			return
		}

		foundOrganization = &apiResponse.Organizations[0]
	} else {
		errors.AddValidationDiagnostic(&resp.Diagnostics, "Organization Lookup", "ID or Name must be provided")
		return
	}

	// Map api response to State
	data.MapFromApi(foundOrganization)

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
