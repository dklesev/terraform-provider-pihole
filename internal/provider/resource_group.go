// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &GroupResource{}
	_ resource.ResourceWithImportState = &GroupResource{}
)

// NewGroupResource creates a new group resource.
func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

// GroupResource defines the resource implementation.
type GroupResource struct {
	client *client.Client
}

// GroupResourceModel describes the resource data model.
type GroupResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Description  types.String `tfsdk:"description"`
	DateAdded    types.Int64  `tfsdk:"date_added"`
	DateModified types.Int64  `tfsdk:"date_modified"`
}

func (r *GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole group for organizing clients and domain rules.",
		MarkdownDescription: `
Manages a Pi-hole group for organizing clients and domain rules.

Groups allow you to apply different blocking rules to different sets of clients.

## Example Usage

` + "```hcl" + `
resource "pihole_group" "trusted_devices" {
  name        = "trusted_devices"
  enabled     = true
  description = "Devices with relaxed ad blocking"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the group in Pi-hole.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The unique name of the group.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the group is enabled. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"description": schema.StringAttribute{
				Description: "A description of the group.",
				Optional:    true,
			},
			"date_added": schema.Int64Attribute{
				Description: "Unix timestamp when the group was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"date_modified": schema.Int64Attribute{
				Description: "Unix timestamp when the group was last modified.",
				Computed:    true,
			},
		},
	}
}

func (r *GroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating group", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	group := &client.Group{
		Name:        data.Name.ValueString(),
		Enabled:     data.Enabled.ValueBool(),
		Description: data.Description.ValueString(),
	}

	created, err := r.client.CreateGroup(ctx, group)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating group",
			fmt.Sprintf("Could not create group %s: %s", data.Name.ValueString(), err.Error()),
		)
		return
	}

	r.mapGroupToModel(created, &data)

	tflog.Debug(ctx, "Created group", map[string]interface{}{
		"id":   created.ID,
		"name": created.Name,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading group", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	group, err := r.client.GetGroup(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading group",
			fmt.Sprintf("Could not read group %s: %s", data.Name.ValueString(), err.Error()),
		)
		return
	}

	if group == nil {
		// Group was deleted outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapGroupToModel(group, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GroupResourceModel
	var state GroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating group", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	group := &client.Group{
		Name:        data.Name.ValueString(),
		Enabled:     data.Enabled.ValueBool(),
		Description: data.Description.ValueString(),
	}

	updated, err := r.client.UpdateGroup(ctx, state.Name.ValueString(), group)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating group",
			fmt.Sprintf("Could not update group %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}

	r.mapGroupToModel(updated, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting group", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	err := r.client.DeleteGroup(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting group",
			fmt.Sprintf("Could not delete group %s: %s", data.Name.ValueString(), err.Error()),
		)
		return
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by name
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

func (r *GroupResource) mapGroupToModel(group *client.Group, data *GroupResourceModel) {
	data.ID = types.Int64Value(group.ID)
	data.Name = types.StringValue(group.Name)
	data.Enabled = types.BoolValue(group.Enabled)

	if group.Description != "" {
		data.Description = types.StringValue(group.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.DateAdded = types.Int64Value(group.DateAdded)
	data.DateModified = types.Int64Value(group.DateModified)
}
