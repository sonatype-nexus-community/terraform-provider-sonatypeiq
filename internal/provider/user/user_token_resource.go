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
	"net/http"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// userTokenResource is the resource implementation.
type userTokenResource struct {
	common.BaseResource
}

// NewUserTokenResource is a helper function to simplify the provider implementation.
func NewUserTokenResource() resource.Resource {
	return &userTokenResource{}
}

// Metadata returns the resource type name.
func (r *userTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_token"
}

// Schema defines the schema for the resource.
func (r *userTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage the authenticated Users' User Token",
		Attributes: map[string]schema.Attribute{
			"generated_at": schema.StringAttribute{
				Description: "A field used to record the request for User Token generation. Changing this will re-generate the User Token. It's use is purely to allow you to determine when to rotate your User Token.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			// "username": schema.StringAttribute{
			// 	Description: "Username of the current User for which this User Token relates",
			// 	Computed:    true,
			// },
			// "realm": schema.StringAttribute{
			// 	Description: "The authentication Realm that this User belongs to",
			// 	Computed:    true,
			// },
			"user_code": schema.StringAttribute{
				Description: "User Code portion of the User Token",
				Computed:    true,
			},
			"pass_code": schema.StringAttribute{
				Description: "Pass Code portion of the User Token",
				Computed:    true,
				Sensitive:   true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *userTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan model.UserTokenModelResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create User Token
	r.createUserToken(&plan, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *userTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.UserTokenModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Check User Token still exists with same User Code
	_, httpResponse, err := r.Client.UserTokensAPI.GetUserTokenExistsForCurrentUser(ctx).Execute()

	if err != nil {
		if httpResponse.StatusCode == http.StatusNotFound {
			// There is no User Token currently with the Username + Realm combo
			resp.State.RemoveResource(ctx)
			common.HandleApiWarning(
				"No User Token exists for User in Realm",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		} else {
			common.HandleApiError(
				"Error checking User Token",
				&err,
				httpResponse,
				&resp.Diagnostics,
			)
		}
		return
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *userTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan model.UserTokenModelResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to delete existing User Token first , then Create
	r.deleteUserToken(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new User Token
	r.createUserToken(&plan, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *userTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from plan
	var state model.UserTokenModelResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete User Token
	r.deleteUserToken(ctx, &resp.Diagnostics)
}

func (r *userTokenResource) createUserToken(stateModel *model.UserTokenModelResource, ctx context.Context, respDiags *diag.Diagnostics) {
	// Call API to create Organization
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	apiResponse, httpResponse, err := r.Client.UserTokensAPI.CreateUserToken(ctx).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusOK {
		common.HandleApiError(
			"Error creating User Token",
			&err,
			httpResponse,
			respDiags,
		)
		return
	}

	stateModel.MapFromApi(apiResponse)
	stateModel.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}

func (r *userTokenResource) deleteUserToken(ctx context.Context, respDiags *diag.Diagnostics) {
	// Call API to create Organization
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	httpResponse, err := r.Client.UserTokensAPI.DeleteCurrentUserToken(ctx).Execute()

	if err != nil || httpResponse.StatusCode != http.StatusNoContent {
		common.HandleApiError(
			"Error deleting User Token",
			&err,
			httpResponse,
			respDiags,
		)
		return
	}
}
