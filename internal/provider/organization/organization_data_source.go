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
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	"github.com/sonatype-nexus-community/terraform-provider-shared/schema"
	sharedutil "github.com/sonatype-nexus-community/terraform-provider-shared/util"
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
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		d.Auth,
	)

	var org *sonatypeiq.ApiOrganizationDTO
	var r *http.Response
	var err error

	if !data.ID.IsNull() {
		// Lookup By Org ID
		org, r, err = d.Client.OrganizationsAPI.GetOrganization(ctx, data.ID.ValueString()).Execute()
		if err != nil {
			sharederr.HandleAPIError(
				"Unable to Read IQ Organization by ID",
				&err,
				r,
				&resp.Diagnostics,
			)
			return
		}
		if r.StatusCode != 200 {
			sharederr.AddAPIErrorDiagnostic(&resp.Diagnostics, "Read Organization", "Organization", r, err)
			return
		}

	} else if !data.Name.IsNull() {
		// Lookup By Org ID
		var orgs *sonatypeiq.ApiOrganizationListDTO
		get_orgs_req := d.Client.OrganizationsAPI.GetOrganizations(ctx)
		get_orgs_req = get_orgs_req.OrganizationName([]string{data.Name.ValueString()})
		orgs, r, err = get_orgs_req.Execute()
		if err != nil {
			sharederr.HandleAPIError(
				"Unable to Read IQ Organization by Name",
				&err,
				r,
				&resp.Diagnostics,
			)
			return
		}
		if r.StatusCode != 200 {
			sharederr.AddAPIErrorDiagnostic(&resp.Diagnostics, "Read Organizations", "Organization", r, err)
			return
		}
		if len(orgs.Organizations) == 1 {
			org = &orgs.Organizations[0]
		} else if len(orgs.Organizations) > 1 {
			sharederr.AddValidationDiagnostic(&resp.Diagnostics, "Organization", "More than one Organization matched the supplied name")
			return
		}
	} else {
		sharederr.AddValidationDiagnostic(&resp.Diagnostics, "Organization", "ID or Name must be provided")
		return
	}

	if org == nil {
		sharederr.AddValidationDiagnostic(&resp.Diagnostics, "Organization", "No Organization found with the provided ID or Name")
		return
	}

	var parentOrgId = types.StringNull()
	if org.ParentOrganizationId != nil {
		tflog.Debug(ctx, fmt.Sprintf("Parent Org Id is %s", *org.ParentOrganizationId))
		parentOrgId = sharedutil.StringPtrToValue(org.ParentOrganizationId)
	}
	om := model.OrganizationModel{
		ID:                    sharedutil.StringPtrToValue(org.Id),
		Name:                  sharedutil.StringPtrToValue(org.Name),
		ParentOrganiziationId: parentOrgId,
		Tags:                  nil,
	}
	for _, tag := range org.Tags {
		om.Tags = append(om.Tags, model.TagModel{
			ID:          sharedutil.StringPtrToValue(tag.Id),
			Name:        sharedutil.StringPtrToValue(tag.Name),
			Description: sharedutil.StringPtrToValue(tag.Description),
			Color:       sharedutil.StringPtrToValue(tag.Color),
		})
	}

	// Set state
	diags := resp.State.Set(ctx, &om)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
