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
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	"github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &systemConfigDataSource{}
	_ datasource.DataSourceWithConfigure = &systemConfigDataSource{}
)

// SystemConfigDataSource is a helper function to simplify the provider implementation.
func SystemConfigDataSource() datasource.DataSource {
	return &systemConfigDataSource{}
}

// systemConfigDataSource is the data source implementation.
type systemConfigDataSource struct {
	common.BaseDataSource
}

// Metadata returns the data source type name.
func (d *systemConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_config"
}

// Schema defines the schema for the data source.
func (d *systemConfigDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Description: "Use this data source to get System Configuration",
		Attributes: map[string]tfschema.Attribute{
			"id":             schema.DataSourceComputedString("The ID of this resource."),
			"base_url":       schema.DataSourceComputedString("Base URL for Sonatype IQ Server"),
			"force_base_url": schema.DataSourceComputedBool("Should the Base URL be forced?"),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *systemConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.SystemConfigModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	apiResponse, httpResponse, err := d.Client.ConfigurationAPI.GetConfiguration(d.AuthContext(ctx)).Property([]sonatypeiq.SystemConfigProperty{"baseUrl", "forceBaseUrl"}).Execute()

	if err != nil {
		errors.HandleAPIError(
			common.ERR_FAILED_READING_SYSTEM_CONFIG,
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.AddAPIErrorDiagnostic(&resp.Diagnostics, "read", "System Configuration", httpResponse, err)
		return
	}

	// Assign Response Data to State
	data.ID = types.StringValue("system-configuration")
	data.MapFromApi(apiResponse)

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
