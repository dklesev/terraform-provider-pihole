// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ClientResource{}
	_ resource.ResourceWithImportState = &ClientResource{}
)

func NewClientResource() resource.Resource {
	return &ClientResource{}
}

type ClientResource struct {
	client *client.Client
}

type ClientResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Client       types.String `tfsdk:"client"`
	Comment      types.String `tfsdk:"comment"`
	Groups       types.List   `tfsdk:"groups"`
	DateAdded    types.Int64  `tfsdk:"date_added"`
	DateModified types.Int64  `tfsdk:"date_modified"`
}

func (r *ClientResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

func (r *ClientResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole client configuration.",
		MarkdownDescription: `
Manages a Pi-hole client configuration.

Clients can be identified by IP address, MAC address, hostname, CIDR subnet, or interface name.

## Example Usage

### By IP Address

` + "```hcl" + `
resource "pihole_client" "workstation" {
  client  = "192.168.1.100"
  groups  = [pihole_group.trusted.id]
  comment = "Developer workstation"
}
` + "```" + `

### By MAC Address

` + "```hcl" + `
resource "pihole_client" "laptop" {
  client  = "AA:BB:CC:DD:EE:FF"
  groups  = [pihole_group.trusted.id]
  comment = "My laptop"
}
` + "```" + `

### By Subnet

` + "```hcl" + `
resource "pihole_client" "iot_subnet" {
  client  = "192.168.10.0/24"
  groups  = [pihole_group.iot.id]
  comment = "IoT devices subnet"
}
` + "```" + `

### By Interface

` + "```hcl" + `
resource "pihole_client" "guest_wifi" {
  client  = ":wlan1"
  groups  = [pihole_group.guests.id]
  comment = "Guest WiFi interface"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the client in Pi-hole.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"client": schema.StringAttribute{
				Description: "The client identifier (IP, MAC, hostname, CIDR subnet, or interface prefixed with ':').",
				Required:    true,
			},
			"comment": schema.StringAttribute{
				Description: "A comment describing the client.",
				Optional:    true,
			},
			"groups": schema.ListAttribute{
				Description: "List of group IDs this client belongs to. Default group ID is 0.",
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"date_added": schema.Int64Attribute{
				Description: "Unix timestamp when the client was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"date_modified": schema.Int64Attribute{
				Description: "Unix timestamp when the client was last modified.",
				Computed:    true,
			},
		},
	}
}

func (r *ClientResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ClientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ClientResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating client", map[string]interface{}{
		"client": data.Client.ValueString(),
	})

	var groups []int64
	if !data.Groups.IsNull() && !data.Groups.IsUnknown() {
		resp.Diagnostics.Append(data.Groups.ElementsAs(ctx, &groups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	piholeClient := &client.PiholeClient{
		Client:  data.Client.ValueString(),
		Comment: data.Comment.ValueString(),
		Groups:  groups,
	}

	created, err := r.client.CreateClient(ctx, piholeClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating client",
			fmt.Sprintf("Could not create client %s: %s", data.Client.ValueString(), err.Error()),
		)
		return
	}

	r.mapClientToModel(ctx, created, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClientResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	piholeClient, err := r.client.GetClient(ctx, data.Client.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading client",
			fmt.Sprintf("Could not read client %s: %s", data.Client.ValueString(), err.Error()),
		)
		return
	}

	if piholeClient == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapClientToModel(ctx, piholeClient, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ClientResourceModel
	var state ClientResourceModel

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

	piholeClient := &client.PiholeClient{
		Client:  data.Client.ValueString(),
		Comment: data.Comment.ValueString(),
		Groups:  groups,
	}

	updated, err := r.client.UpdateClient(ctx, state.Client.ValueString(), piholeClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating client",
			fmt.Sprintf("Could not update client %s: %s", state.Client.ValueString(), err.Error()),
		)
		return
	}

	r.mapClientToModel(ctx, updated, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ClientResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteClient(ctx, data.Client.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting client",
			fmt.Sprintf("Could not delete client %s: %s", data.Client.ValueString(), err.Error()),
		)
		return
	}
}

func (r *ClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("client"), req, resp)
}

func (r *ClientResource) mapClientToModel(ctx context.Context, piholeClient *client.PiholeClient, data *ClientResourceModel, diags *diag.Diagnostics) {
	data.ID = types.Int64Value(piholeClient.ID)
	data.Client = types.StringValue(piholeClient.Client)

	if piholeClient.Comment != "" {
		data.Comment = types.StringValue(piholeClient.Comment)
	} else {
		data.Comment = types.StringNull()
	}

	if len(piholeClient.Groups) > 0 {
		groupsList, d := types.ListValueFrom(ctx, types.Int64Type, piholeClient.Groups)
		diags.Append(d...)
		data.Groups = groupsList
	} else {
		data.Groups = types.ListNull(types.Int64Type)
	}

	data.DateAdded = types.Int64Value(piholeClient.DateAdded)
	data.DateModified = types.Int64Value(piholeClient.DateModified)
}
