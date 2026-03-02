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
	"terraform-provider-sonatypeiq/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
	sharedutil "github.com/sonatype-nexus-community/terraform-provider-shared/util"
)

type organizationModelResouce struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	ParentOrganiziationId types.String `tfsdk:"parent_organization_id"`
	// Tags                  types.List   `tfsdk:"tags"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// organizationResource is the resource implementation.
type organizationResource struct {
	common.BaseResource
}

// NewOrganizationResource is a helper function to simplify the provider implementation.
func NewOrganizationResource() resource.Resource {
	return &organizationResource{}
}

// Metadata returns the resource type name.
func (r *organizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

// Schema defines the schema for the resource.
func (r *organizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this resource to manage Organizations",
		Attributes: map[string]schema.Attribute{
			"id":                     sharedrschema.ResourceComputedString("Internal ID of the Organization"),
			"name":                   sharedrschema.ResourceComputedOptionalString("Name of the Organization"),
			"parent_organization_id": sharedrschema.ResourceComputedOptionalString("Internal ID of the Parent Organization if this Organization has a Parent Organization"),
			// "tags": schema.ListNestedAttribute{
			// 	Optional: true,
			// 	NestedObject: schema.NestedAttributeObject{
			// 		Attributes: tagSchemaObjectAttributes,
			// 	},
			// },
			"last_updated": sharedrschema.ResourceComputedString("String representation of the date/time the resource was last changed by Terraform"),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan organizationModelResouce
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		print("*** ERRORS ***")
		return
	}

	tflog.Debug(ctx, "Preparing to create Organization", map[string]interface{}{"orgConfig": fmt.Sprintf("%+v", plan)})

	// Call API to create Organization
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	organization_request := r.Client.OrganizationsAPI.AddOrganization(ctx)
	orgDto := sonatypeiq.ApiOrganizationDTO{
		Name:                 sharedutil.StringToPtr(plan.Name.ValueString()),
		ParentOrganizationId: sharedutil.StringToPtr(plan.ParentOrganiziationId.ValueString()),
	}

	// if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() && len(plan.Tags.Elements()) > 0 {
	// 	tflog.Debug(ctx, "Adding Tag to Organization Create Request...")

	// 	tags := make([]tagModel, len(plan.Tags.Elements()))

	// 	for _, tag := range tags {
	// 		orgDto.Tags = append(orgDto.Tags, sonatypeiq.ApiTagDTO{
	// 			Id:          tag.ID.ValueStringPointer(),
	// 			Name:        tag.Name.ValueStringPointer(),
	// 			Description: tag.Description.ValueStringPointer(),
	// 			Color:       tag.Color.ValueStringPointer(),
	// 		})
	// 	}
	// }

	organization_request = organization_request.ApiOrganizationDTO(orgDto)

	organization, api_response, err := organization_request.Execute()

	// Call API
	if err != nil {
		sharederr.HandleAPIError("Error creating Organization", &err, api_response, &resp.Diagnostics)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = sharedutil.StringPtrToValue(organization.Id)
	plan.Name = sharedutil.StringPtrToValue(organization.Name)
	plan.ParentOrganiziationId = sharedutil.StringPtrToValue(organization.ParentOrganizationId)
	// plan.Tags = []tagModel{}
	// for _, tagDto := range organization.Tags {
	// 	plan.Tags = append(plan.Tags, tagModel{
	// 		ID:          types.StringValue(*tagDto.Id),
	// 		Name:        types.StringValue(*tagDto.Name),
	// 		Description: types.StringValue(*tagDto.Description),
	// 		Color:       types.StringValue(*tagDto.Color),
	// 	})
	// }
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Finally, set the state
	tflog.Debug(ctx, "Storing certificate request info into the state")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationModelResouce

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	// Get refreshed Organization from IQ
	organization, httpResponse, err := r.Client.OrganizationsAPI.GetOrganization(ctx, state.ID.ValueString()).Execute()

	if err != nil {
		sharederr.HandleAPIError("Error reading Organization", &err, httpResponse, &resp.Diagnostics)
		return
	}

	// Overwrite items with refreshed state
	state.ID = sharedutil.StringPtrToValue(organization.Id)
	state.Name = sharedutil.StringPtrToValue(organization.Name)
	state.ParentOrganiziationId = sharedutil.StringPtrToValue(organization.ParentOrganizationId)

	// if len(organization.Tags) > 0 {
	// 	tflog.Debug(ctx, "Adding Tag to Organization Read response...")

	// 	tags := []attr.Value{}

	// 	for _, tag := range organization.Tags {
	// 		tag := map[string]attr.Value{
	// 			"id":          types.StringValue(tag.GetId()),
	// 			"name":        types.StringValue(tag.GetName()),
	// 			"description": types.StringValue(tag.GetDescription()),
	// 			"color":       types.StringValue(tag.GetColor()),
	// 		}

	// 		tagObj, _ := types.ObjectValue(tagObjectMemberTypes, tag)
	// 		tags = append(tags, tagObj)
	// 	}

	// 	state.Tags, _ = types.ListValue(types.ObjectType{AttrTypes: tagObjectMemberTypes}, tags)
	// }

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No Update API
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationModelResouce
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make Delete API Call
	ctx = context.WithValue(
		ctx,
		sonatypeiq.ContextBasicAuth,
		r.Auth,
	)

	api_response, err := r.Client.OrganizationsAPI.DeleteOrganization(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		sharederr.HandleAPIError("Error deleting Organization", &err, api_response, &resp.Diagnostics)
		return
	}
}

func (r *organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
