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

package system

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
	_ datasource.DataSource              = &systemConfigDataSource{}
	_ datasource.DataSourceWithConfigure = &systemConfigDataSource{}
)

// ConfigSamlDataSource is a helper function to simplify the provider implementation.
func ConfigSamlDataSource() datasource.DataSource {
	return &configSamlDataSource{}
}

// configSamlDataSource is the data source implementation.
type configSamlDataSource struct {
	common.BaseDataSource
}

// Metadata returns the data source type name.
func (d *configSamlDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_saml"
}

// Schema defines the schema for the data source.
func (d *configSamlDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Description: "Use this data source to get SAML Metadata for Sonatype IQ Server",
		Attributes: map[string]tfschema.Attribute{
			"id":            schema.DataSourceComputedString("The ID of this resource."),
			"saml_metadata": schema.DataSourceComputedString("SAML Metadata for Sonatype IQ Server"),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *configSamlDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.SecuritySamlMetadataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := d.Client.ConfigSAMLAPI.GetMetadata(d.AuthContext(ctx)).Execute()

	if err != nil && httpResponse.StatusCode != http.StatusNotFound {
		errors.HandleAPIError(
			common.ERR_FAILED_READING_SAML_METADATA,
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK && httpResponse.StatusCode != http.StatusNotFound {
		errors.AddAPIErrorDiagnostic(&resp.Diagnostics, "read", "SAML Metadata", httpResponse, err)
		return
	}

	data.ID = types.StringValue("saml-metadata")
	if httpResponse.StatusCode == http.StatusOK {
		data.SamlMetadata = types.StringValue(apiResponse)
	} else {
		data.SamlMetadata = types.StringNull()
	}

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
