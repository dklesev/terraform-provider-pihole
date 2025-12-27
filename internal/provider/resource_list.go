// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

var (
	_ resource.Resource                = &ListResource{}
	_ resource.ResourceWithImportState = &ListResource{}
)

func NewListResource() resource.Resource {
	return &ListResource{}
}

type ListResource struct {
	client *client.Client
}

type ListResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Address      types.String `tfsdk:"address"`
	Type         types.String `tfsdk:"type"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Comment      types.String `tfsdk:"comment"`
	Groups       types.Set    `tfsdk:"groups"`
	DateAdded    types.Int64  `tfsdk:"date_added"`
	DateModified types.Int64  `tfsdk:"date_modified"`
	Number       types.Int64  `tfsdk:"number"`
	Status       types.Int64  `tfsdk:"status"`
}

func (r *ListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_list"
}

func (r *ListResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole blocklist or allowlist subscription.",
		MarkdownDescription: `
Manages a Pi-hole blocklist or allowlist subscription.

Lists are external URLs containing domains that Pi-hole will block or allow.

## Example Usage

### Blocklist

` + "```hcl" + `
resource "pihole_list" "hagezi_pro" {
  address = "https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/adblock/pro.txt"
  type    = "block"
  enabled = true
  comment = "Hagezi Pro blocklist"
}
` + "```" + `

### Allowlist

` + "```hcl" + `
resource "pihole_list" "whitelist" {
  address = "https://example.com/allowlist.txt"
  type    = "allow"
  enabled = true
  comment = "Custom allowlist"
}
` + "```" + `

## Import

Lists can be imported using the format ` + "`type/address`" + `:

` + "```shell" + `
terraform import pihole_list.example block/https://example.com/blocklist.txt
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the list in Pi-hole.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"address": schema.StringAttribute{
				Description: "The URL of the list.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of list: 'block' or 'allow'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("block", "allow"),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the list is enabled. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"comment": schema.StringAttribute{
				Description: "A comment describing the list.",
				Optional:    true,
			},
			"groups": schema.SetAttribute{
				Description: "List of group IDs this list applies to. Default group ID is 0.",
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"date_added": schema.Int64Attribute{
				Description: "Unix timestamp when the list was added.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"date_modified": schema.Int64Attribute{
				Description: "Unix timestamp when the list was last modified.",
				Computed:    true,
			},
			"number": schema.Int64Attribute{
				Description: "Number of domains in the list.",
				Computed:    true,
			},
			"status": schema.Int64Attribute{
				Description: "Download status of the list.",
				Computed:    true,
			},
		},
	}
}

func (r *ListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *ListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ListResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating list", map[string]interface{}{
		"address": data.Address.ValueString(),
		"type":    data.Type.ValueString(),
	})

	var groups []int64
	if !data.Groups.IsNull() && !data.Groups.IsUnknown() {
		resp.Diagnostics.Append(data.Groups.ElementsAs(ctx, &groups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	list := &client.List{
		Address: data.Address.ValueString(),
		Type:    data.Type.ValueString(),
		Enabled: data.Enabled.ValueBool(),
		Comment: data.Comment.ValueString(),
		Groups:  groups,
	}

	created, err := r.client.CreateList(ctx, list)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating list",
			fmt.Sprintf("Could not create list %s: %s", data.Address.ValueString(), err.Error()),
		)
		return
	}

	r.mapListToModel(ctx, created, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ListResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.client.GetList(ctx, data.Type.ValueString(), data.Address.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading list",
			fmt.Sprintf("Could not read list %s: %s", data.Address.ValueString(), err.Error()),
		)
		return
	}

	if list == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapListToModel(ctx, list, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ListResourceModel
	var state ListResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var groups []int64
	if !data.Groups.IsNull() && !data.Groups.IsUnknown() {
		resp.Diagnostics.Append(data.Groups.ElementsAs(ctx, &groups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	list := &client.List{
		Address: data.Address.ValueString(),
		Type:    data.Type.ValueString(),
		Enabled: data.Enabled.ValueBool(),
		Comment: data.Comment.ValueString(),
		Groups:  groups,
	}

	updated, err := r.client.UpdateList(ctx, state.Type.ValueString(), state.Address.ValueString(), list)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating list",
			fmt.Sprintf("Could not update list %s: %s", state.Address.ValueString(), err.Error()),
		)
		return
	}

	r.mapListToModel(ctx, updated, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ListResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteList(ctx, data.Type.ValueString(), data.Address.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting list",
			fmt.Sprintf("Could not delete list %s: %s", data.Address.ValueString(), err.Error()),
		)
		return
	}
}

func (r *ListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: type/address
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format: type/address (e.g., block/https://example.com/list.txt)",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("address"), parts[1])...)
}

func (r *ListResource) mapListToModel(ctx context.Context, list *client.List, data *ListResourceModel, diags *diag.Diagnostics) {
	data.ID = types.Int64Value(list.ID)
	data.Address = types.StringValue(list.Address)
	data.Type = types.StringValue(list.Type)
	data.Enabled = types.BoolValue(list.Enabled)

	if list.Comment != "" {
		data.Comment = types.StringValue(list.Comment)
	} else {
		data.Comment = types.StringNull()
	}

	if len(list.Groups) > 0 {
		groupsList, d := types.SetValueFrom(ctx, types.Int64Type, list.Groups)
		diags.Append(d...)
		data.Groups = groupsList
	} else {
		data.Groups = types.SetNull(types.Int64Type)
	}

	data.DateAdded = types.Int64Value(list.DateAdded)
	data.DateModified = types.Int64Value(list.DateModified)
	data.Number = types.Int64Value(list.Number)
	data.Status = types.Int64Value(int64(list.Status))
}
