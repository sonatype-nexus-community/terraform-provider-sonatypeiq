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
	"crypto/sha1"
	"fmt"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
	sharedutil "github.com/sonatype-nexus-community/terraform-provider-shared/util"
)

// userResource is the resource implementation.
type userResource struct {
	common.BaseResource
}

type userModelResource struct {
	ID          types.String `tfsdk:"id"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	FirstName   types.String `tfsdk:"first_name"`
	LastName    types.String `tfsdk:"last_name"`
	Email       types.String `tfsdk:"email"`
	Realm       types.String `tfsdk:"realm"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (u *userModelResource) GenerateID() {
	hash := sha1.New()
	hash.Write([]byte(u.Username.ValueString() + u.Realm.ValueString()))
	u.ID = types.StringValue(fmt.Sprintf("%x", hash.Sum(nil)))
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
			"id":         sharedrschema.ResourceComputedString("The user ID"),
			"username":   sharedrschema.ResourceRequiredString("Username used to log in to Sonatype IQ Server"),
			"password":   sharedrschema.ResourceSensitiveString("Password used to log in to Sonatype IQ Server"),
			"first_name": sharedrschema.ResourceRequiredString("Users first name"),
			"last_name":  sharedrschema.ResourceRequiredString("Users last name"),
			"email":      sharedrschema.ResourceRequiredString("Users email address"),
			"realm": sharedrschema.ResourceOptionalStringWithDefault(
				fmt.Sprintf("Realm the User belongs to. Only '%s' is supported at this time.", common.DEFAULT_USER_REALM),
				common.DEFAULT_USER_REALM,
			),
			"last_updated": sharedrschema.ResourceComputedString("String representation of the date/time the resource was last changed by Terraform"),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan userModelResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validation
	if plan.Password.IsNull() || plan.Password.IsUnknown() {
		sharederr.AddValidationDiagnostic(&resp.Diagnostics, "Password", "Password is required when creating a User")
		return
	}
	if !plan.Realm.Equal(types.StringValue(common.DEFAULT_USER_REALM)) {
		sharederr.AddValidationDiagnostic(&resp.Diagnostics, "Realm", fmt.Sprintf("Only the '%s' Realm is supported currently.", common.DEFAULT_USER_REALM))
		return
	}

	// Call API to create Organization
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	user_request := r.Client.UsersAPI.Add(ctx)
	user_config := sonatypeiq.ApiUserDTO{
		Username:  sharedutil.StringToPtr(plan.Username.ValueString()),
		Password:  sharedutil.StringToPtr(plan.Password.ValueString()),
		FirstName: sharedutil.StringToPtr(plan.FirstName.ValueString()),
		LastName:  sharedutil.StringToPtr(plan.LastName.ValueString()),
		Email:     sharedutil.StringToPtr(plan.Email.ValueString()),
		Realm:     sharedutil.StringToPtr(plan.Realm.ValueString()),
	}
	user_request = user_request.ApiUserDTO(user_config)
	api_response, err := user_request.Execute()

	// Call API
	if err != nil || api_response.StatusCode != 204 {
		sharederr.HandleAPIError(
			"Error creating User",
			&err,
			api_response,
			&resp.Diagnostics,
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.GenerateID()
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userModelResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Lookup System Configuration
	user, api_response, err := r.Client.UsersAPI.Get1(ctx, state.Username.ValueString()).Execute()

	if err != nil || api_response.StatusCode != 200 {
		sharederr.HandleAPIError(
			"Error reading User",
			&err,
			api_response,
			&resp.Diagnostics,
		)
		return
	}

	state.FirstName = sharedutil.StringPtrToValue(user.FirstName)
	state.LastName = sharedutil.StringPtrToValue(user.LastName)
	state.Email = sharedutil.StringPtrToValue(user.Email)
	state.Realm = sharedutil.StringPtrToValue(user.Realm)

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan userModelResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state userModelResource
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validation
	if !plan.Password.IsNull() || !plan.Password.IsUnknown() {
		sharederr.AddValidationDiagnostic(&resp.Diagnostics, "Password", "Changing User Password is not supported by Sonatype IQ Server")
	}
	if !plan.Realm.Equal(types.StringValue("Internal")) {
		sharederr.AddValidationDiagnostic(&resp.Diagnostics, "Realm", fmt.Sprintf("Only the '%s' Realm is supported currently.", common.DEFAULT_USER_REALM))
		return
	}

	// Call API to create Organization
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	user_request := r.Client.UsersAPI.Update(ctx, state.Username.ValueString())
	user_config := sonatypeiq.ApiUserDTO{
		Username:  sharedutil.StringToPtr(plan.Username.ValueString()),
		FirstName: sharedutil.StringToPtr(plan.FirstName.ValueString()),
		LastName:  sharedutil.StringToPtr(plan.LastName.ValueString()),
		Email:     sharedutil.StringToPtr(plan.Email.ValueString()),
		Realm:     sharedutil.StringToPtr(plan.Realm.ValueString()),
	}
	// Changing User Password not supported by IQ
	// if !plan.Password.IsNull() {
	// 	user_config.Password = sharedutil.StringToPtr(plan.Password.ValueString())
	// }
	user_request = user_request.ApiUserDTO(user_config)
	user, api_response, err := user_request.Execute()

	// Call API
	if err != nil || api_response.StatusCode != 200 {
		sharederr.HandleAPIError(
			"Error updating User",
			&err,
			api_response,
			&resp.Diagnostics,
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.Username = sharedutil.StringPtrToValue(user.Username)
	plan.FirstName = sharedutil.StringPtrToValue(user.FirstName)
	plan.LastName = sharedutil.StringPtrToValue(user.LastName)
	plan.Email = sharedutil.StringPtrToValue(user.Email)
	plan.Realm = sharedutil.StringPtrToValue(user.Realm)
	plan.GenerateID()
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from plan
	var plan userModelResource
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to create Organization
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Call Delete API
	api_response, err := r.Client.UsersAPI.Delete1(ctx, plan.Username.ValueString()).Execute()
	if err != nil || api_response.StatusCode != 204 {
		sharederr.HandleAPIError(
			"Error deleting User",
			&err,
			api_response,
			&resp.Diagnostics,
		)
		return
	}
}
