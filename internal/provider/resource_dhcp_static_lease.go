// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &DHCPStaticLeaseResource{}
	_ resource.ResourceWithImportState = &DHCPStaticLeaseResource{}
)

func NewDHCPStaticLeaseResource() resource.Resource {
	return &DHCPStaticLeaseResource{}
}

type DHCPStaticLeaseResource struct {
	client *client.Client
}

type DHCPStaticLeaseResourceModel struct {
	ID       types.String `tfsdk:"id"`
	MAC      types.String `tfsdk:"mac"`
	IP       types.String `tfsdk:"ip"`
	Hostname types.String `tfsdk:"hostname"`
}

func (r *DHCPStaticLeaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_static_lease"
}

func (r *DHCPStaticLeaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole DHCP static lease.",
		MarkdownDescription: `
Manages a DHCP static lease (MAC -> IP reservation) in Pi-hole.

## Example Usage

` + "```hcl" + `
resource "pihole_dhcp_static_lease" "server" {
  mac      = "AA:BB:CC:DD:EE:FF"
  ip       = "192.168.1.100"
  hostname = "server"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier.",
			},
			"mac": schema.StringAttribute{
				Required:    true,
				Description: "The MAC address of the device.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip": schema.StringAttribute{
				Required:    true,
				Description: "The reserved IP address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				Required:    true,
				Description: "The hostname for the device.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *DHCPStaticLeaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData))
		return
	}
	r.client = c
}

func (r *DHCPStaticLeaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DHCPStaticLeaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Format: "MAC,IP,hostname"
	value := fmt.Sprintf("%s,%s,%s", data.MAC.ValueString(), data.IP.ValueString(), data.Hostname.ValueString())
	tflog.Debug(ctx, "Creating DHCP static lease", map[string]interface{}{"value": value})

	if err := r.client.AddConfigArrayItem(ctx, "dhcp/hosts", value); err != nil {
		resp.Diagnostics.AddError("Error adding DHCP static lease", err.Error())
		return
	}

	data.ID = types.StringValue(value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DHCPStaticLeaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DHCPStaticLeaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	value := fmt.Sprintf("%s,%s,%s", data.MAC.ValueString(), data.IP.ValueString(), data.Hostname.ValueString())

	config, err := r.client.GetDHCPConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading DHCP config", err.Error())
		return
	}

	found := false
	for _, h := range config.Hosts {
		if h == value {
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DHCPStaticLeaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Changes require replacement")
}

func (r *DHCPStaticLeaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DHCPStaticLeaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	value := fmt.Sprintf("%s,%s,%s", data.MAC.ValueString(), data.IP.ValueString(), data.Hostname.ValueString())
	tflog.Debug(ctx, "Deleting DHCP static lease", map[string]interface{}{"value": value})

	if err := r.client.DeleteConfigArrayItem(ctx, "dhcp/hosts", value); err != nil {
		resp.Diagnostics.AddError("Error deleting DHCP static lease", err.Error())
		return
	}
}

func (r *DHCPStaticLeaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "MAC,IP,hostname"
	parts := strings.SplitN(req.ID, ",", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: 'MAC,IP,hostname'")
		return
	}

	data := DHCPStaticLeaseResourceModel{
		ID:       types.StringValue(req.ID),
		MAC:      types.StringValue(parts[0]),
		IP:       types.StringValue(parts[1]),
		Hostname: types.StringValue(parts[2]),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
