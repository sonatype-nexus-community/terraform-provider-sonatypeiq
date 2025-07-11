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
	"io"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
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
			"id": schema.StringAttribute{
				Computed: true,
			},
			"username": schema.StringAttribute{
				Description: "Username used to log in to Sonatype IQ Server",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password used to log in to Sonatype IQ Server",
				Optional:    true,
				Sensitive:   true,
			},
			"first_name": schema.StringAttribute{
				Description: "Users first name",
				Required:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Users last name",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "Users email address",
				Required:    true,
			},
			"realm": schema.StringAttribute{
				Description: fmt.Sprintf("Realm the User belongs to. Only '%s' is supported at this time.", common.DEFAULT_USER_REALM),
				Default:     stringdefault.StaticString(common.DEFAULT_USER_REALM),
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
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
		resp.Diagnostics.AddError(
			"No Password supplied",
			"Password is required when creating a User",
		)
		return
	}
	if !plan.Realm.Equal(types.StringValue(common.DEFAULT_USER_REALM)) {
		resp.Diagnostics.AddError(
			"Unsupported Realm",
			fmt.Sprintf("Only the '%s' Realm is supported currently.", common.DEFAULT_USER_REALM),
		)
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
		Username:  plan.Username.ValueStringPointer(),
		Password:  plan.Password.ValueStringPointer(),
		FirstName: plan.FirstName.ValueStringPointer(),
		LastName:  plan.LastName.ValueStringPointer(),
		Email:     plan.Email.ValueStringPointer(),
		Realm:     plan.Realm.ValueStringPointer(),
	}
	user_request = user_request.ApiUserDTO(user_config)
	api_response, err := user_request.Execute()

	// Call API
	if err != nil || api_response.StatusCode != 204 {
		error_body, _ := io.ReadAll(api_response.Body)
		resp.Diagnostics.AddError(
			"Error creating User",
			"Could not create User, unexpected error: "+api_response.Status+": "+string(error_body),
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
		error_body, _ := io.ReadAll(api_response.Body)
		resp.Diagnostics.AddError(
			"Error reading User",
			"Could not read User, unexpected error: "+api_response.Status+": "+string(error_body),
		)
		return
	}

	state.FirstName = types.StringValue(*user.FirstName)
	state.LastName = types.StringValue(*user.LastName)
	state.Email = types.StringValue(*user.Email)
	state.Realm = types.StringValue(*user.Realm)

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
		resp.Diagnostics.AddWarning(
			"Cannot change User Password - will be skipped",
			"Changing User Password is not supported by Sonatype IQ Server",
		)
	}
	if !plan.Realm.Equal(types.StringValue("Internal")) {
		resp.Diagnostics.AddError(
			"Unsupported Realm",
			fmt.Sprintf("Only the '%s' Realm is supported currently.", common.DEFAULT_USER_REALM),
		)
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
		Username:  plan.Username.ValueStringPointer(),
		FirstName: plan.FirstName.ValueStringPointer(),
		LastName:  plan.LastName.ValueStringPointer(),
		Email:     plan.Email.ValueStringPointer(),
		Realm:     plan.Realm.ValueStringPointer(),
	}
	// Changing User Password not supported by IQ
	// if !plan.Password.IsNull() {
	// 	user_config.Password = plan.Password.ValueStringPointer()
	// }
	user_request = user_request.ApiUserDTO(user_config)
	user, api_response, err := user_request.Execute()

	// Call API
	if err != nil || api_response.StatusCode != 200 {
		error_body, _ := io.ReadAll(api_response.Body)
		resp.Diagnostics.AddError(
			"Error updating User",
			"Could not update User, unexpected error: "+api_response.Status+": "+string(error_body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.Username = types.StringValue(*user.Username)
	plan.FirstName = types.StringValue(*user.FirstName)
	plan.LastName = types.StringValue(*user.LastName)
	plan.Email = types.StringValue(*user.Email)
	plan.Realm = types.StringValue(*user.Realm)
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
		error_body, _ := io.ReadAll(api_response.Body)
		resp.Diagnostics.AddError(
			"Error deleting User",
			"Could not delete User, unexpected error: "+api_response.Status+": "+string(error_body),
		)
		return
	}
}
