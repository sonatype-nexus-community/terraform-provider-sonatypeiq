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

package user

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	"github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// userResource is the resource implementation.
type userResource struct {
	common.BaseResource
}

// NewUserResource is a helper function to simplify the provider implementation.
func NewUserResource() resource.Resource {
	return &userResource{}
}

// Metadata returns the resource type name.
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the resource.
func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to manage Users",
		Attributes: map[string]schema.Attribute{
			"id":         sharedrschema.ResourceComputedString("Internal ID for Terraform State"),
			"username":   sharedrschema.ResourceRequiredString("Username used to log in to Sonatype IQ Server"),
			"password":   sharedrschema.ResourceSensitiveString("Password used to log in to Sonatype IQ Server"),
			"first_name": sharedrschema.ResourceRequiredString("Users first name"),
			"last_name":  sharedrschema.ResourceRequiredString("Users last name"),
			"email":      sharedrschema.ResourceRequiredString("Users email address"),
			"realm": sharedrschema.ResourceStringEnumWithDefault(
				fmt.Sprintf("Realm the User belongs to. Only '%s' is supported at this time.", common.DEFAULT_USER_REALM),
				common.DEFAULT_USER_REALM,
				common.DEFAULT_USER_REALM,
			),
			"last_updated": sharedrschema.ResourceLastUpdated(),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.UserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
		return
	}

	httpResponse, err := r.Client.UsersAPI.Add(r.AuthContext(ctx)).ApiUserDTO(*plan.MapToApi(true)).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error creating User",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusNoContent {
		errors.HandleAPIError(
			"Creation of User was not successful",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	apiResponse := r.doRead(ctx, plan.Username.ValueString(), plan.Realm.ValueString(), &resp.State, &resp.Diagnostics)
	if apiResponse == nil {
		return
	}

	// Update State based on Response
	plan.MapFromApi(apiResponse)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state model.UserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	apiResponse := r.doRead(ctx, state.Username.ValueString(), state.Realm.ValueString(), &resp.State, &resp.Diagnostics)
	if apiResponse == nil {
		return
	}

	// Update State based on Response
	state.MapFromApi(apiResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.UserModel
	var state model.UserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_PLAN, resp.Diagnostics.Errors()))
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	// Validation
	if plan.Password.ValueString() != state.Password.ValueString() {
		sharederr.AddValidationDiagnostic(&resp.Diagnostics, "password", "Changing User Password is not supported by Sonatype IQ Server")
	}
	if plan.Realm.ValueString() != common.USER_REALM_INTERNAL {
		sharederr.AddValidationDiagnostic(&resp.Diagnostics, "realm", fmt.Sprintf("Only the '%s' Realm is supported currently.", common.DEFAULT_USER_REALM))
		return
	}

	apiResponse, httpResponse, err := r.Client.UsersAPI.Update(r.AuthContext(ctx), state.Username.ValueString()).ApiUserDTO(*plan.MapToApi(false)).Execute()

	if err != nil {
		errors.HandleAPIError(
			"Error updating User",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	} else if httpResponse.StatusCode != http.StatusOK {
		errors.HandleAPIError(
			"Updating User was not successful",
			&err,
			httpResponse,
			&resp.Diagnostics,
		)
		return
	}

	// Map response to State
	state.MapFromApi(apiResponse)

	// Update State
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state model.UserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf(common.ERR_TF_GETTING_STATE, resp.Diagnostics.Errors()))
		return
	}

	httpResponse, err := r.Client.UsersAPI.Delete1(r.AuthContext(ctx), state.Username.ValueString()).Realm(state.Realm.ValueString()).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			fmt.Sprintf(common.ERR_USER_DID_NOT_EXIST, state.ID.ValueString()),
			fmt.Sprintf("%v", err),
		)
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "-")
	if len(idParts) < 3 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: user-[REALM]-[USERNAME] - Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("realm"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), idParts[2])...)
}

func (r *userResource) doRead(ctx context.Context, username, realm string, respState *tfsdk.State, respDiags *diag.Diagnostics) *sonatypeiq.ApiUserDTO {
	apiResponse, httpResponse, err := r.Client.UsersAPI.Get1(r.AuthContext(ctx), username).Realm(realm).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			respState.RemoveResource(ctx)
			errors.HandleAPIWarning(
				"User did not exist",
				&err,
				httpResponse,
				respDiags,
			)
		} else {
			errors.HandleAPIError(
				fmt.Sprintf(common.ERR_FAILED_READING_USER_AT_REALM, username, realm),
				&err,
				httpResponse,
				respDiags,
			)
		}
		return nil
	}

	return apiResponse
}
