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
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"

	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"

	b64 "encoding/base64"
)

// systemConfigProductLicenseResource is the resource implementation.
type systemConfigProductLicenseResource struct {
	common.BaseResource
}

// NewConfigProductLicenseResource is a helper function to simplify the provider implementation.
func NewConfigProductLicenseResource() resource.Resource {
	return &systemConfigProductLicenseResource{}
}

// Metadata returns the resource type name.
func (r *systemConfigProductLicenseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_product_license"
}

// Schema defines the schema for the resource.
func (r *systemConfigProductLicenseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configure and Sonatype IQ License",
		Attributes: map[string]schema.Attribute{
			"license_data": schema.StringAttribute{
				Description: "Base64 encoded license data",
				Required:    true,
				Optional:    false,
				Sensitive:   true,
			},
			// Meta
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *systemConfigProductLicenseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan model.ProductLicenseModelResource
	var state = model.ProductLicenseModelResource{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting request data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Do the work
	r.updateProductLicense(
		ctx,
		plan.LicenseData.ValueString(),
		&state,
		&resp.State,
		&resp.Diagnostics,
	)
}

// Read refreshes the Terraform state with the latest data.
func (r *systemConfigProductLicenseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// There is no API to read
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *systemConfigProductLicenseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan & state
	var plan model.ProductLicenseModelResource
	var state model.ProductLicenseModelResource

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting plan data has errors: %v", resp.Diagnostics.Errors()))
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Getting state data has errors: %v", resp.Diagnostics.Errors()))
		return
	}

	// Do the work
	r.updateProductLicense(
		ctx,
		plan.LicenseData.ValueString(),
		&state,
		&resp.State,
		&resp.Diagnostics,
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *systemConfigProductLicenseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Set API Context
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)
	httpResponse, err := r.Client.ProductLicenseAPI.UninstallLicense(ctx).Execute()

	// Handle Error
	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		common.HandleApiError(
			"Error removing Product License",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}
}

func (r *systemConfigProductLicenseResource) updateProductLicense(ctx context.Context, licenseDataBase64 string, stateModel *model.ProductLicenseModelResource, tfState *tfsdk.State, respDiags *diag.Diagnostics) {
	// Get and process Product License Base64 Data
	licenseData, err := b64.StdEncoding.DecodeString(licenseDataBase64)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Supplied License Data was not properly Base64 encoded: %v", err))
		return
	}
	productLicenseFile, err := os.CreateTemp("", "sonatype-product-license")
	productLicenseFileName := productLicenseFile.Name()
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Unable to create temporary file for Product License: %v", err))
		return
	}

	_, err = productLicenseFile.Write(licenseData)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Failed to write Product License to temporary file: %v", err))
		return
	}
	_ = productLicenseFile.Close()

	// Seems we have to close and re-open the file in order for the API Client library to be
	// able to zero > 0 bytes in the file to read
	productLicenseFile, err = os.Open(productLicenseFileName)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Could not open temporary license file: %v", err))
		return
	}
	defer func() { _ = os.Remove(productLicenseFileName) }()

	// Call API to Create
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	httpReponse, err := r.Client.ProductLicenseAPI.InstallLicense(ctx).File(productLicenseFile).Execute()

	// Handle Error
	if err != nil || httpReponse.StatusCode != http.StatusOK {
		common.HandleApiError(
			"Error installing Product License",
			&err,
			httpReponse,
			respDiags,
		)
		return
	}

	stateModel.LicenseData = types.StringValue(licenseDataBase64)
	stateModel.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags := tfState.Set(ctx, stateModel)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return
	}
}
