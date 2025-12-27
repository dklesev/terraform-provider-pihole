// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ConfigDHCPResource{}
	_ resource.ResourceWithImportState = &ConfigDHCPResource{}
)

func NewConfigDHCPResource() resource.Resource {
	return &ConfigDHCPResource{}
}

type ConfigDHCPResource struct {
	client *client.Client
}

type ConfigDHCPResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Active               types.Bool   `tfsdk:"active"`
	Start                types.String `tfsdk:"start"`
	End                  types.String `tfsdk:"end"`
	Router               types.String `tfsdk:"router"`
	Netmask              types.String `tfsdk:"netmask"`
	LeaseTime            types.String `tfsdk:"lease_time"`
	IPv6                 types.Bool   `tfsdk:"ipv6"`
	RapidCommit          types.Bool   `tfsdk:"rapid_commit"`
	MultiDNS             types.Bool   `tfsdk:"multi_dns"`
	Logging              types.Bool   `tfsdk:"logging"`
	IgnoreUnknownClients types.Bool   `tfsdk:"ignore_unknown_clients"`
}

func (r *ConfigDHCPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_dhcp"
}

func (r *ConfigDHCPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Pi-hole DHCP server configuration.",
		MarkdownDescription: `
Manages Pi-hole DHCP server configuration.

## Example Usage

` + "```hcl" + `
resource "pihole_config_dhcp" "settings" {
  active    = true
  start     = "192.168.1.100"
  end       = "192.168.1.200"
  router    = "192.168.1.1"
  netmask   = "255.255.255.0"
  lease_time = "24h"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this resource (always 'dhcp').",
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Enable DHCP server.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"start": schema.StringAttribute{
				Description: "Start of DHCP address range.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"end": schema.StringAttribute{
				Description: "End of DHCP address range.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"router": schema.StringAttribute{
				Description: "Router (gateway) IP address.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"netmask": schema.StringAttribute{
				Description: "Netmask for DHCP.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"lease_time": schema.StringAttribute{
				Description: "DHCP lease time (e.g., '24h', '1d').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"ipv6": schema.BoolAttribute{
				Description: "Enable IPv6 DHCP (DHCPv6).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"rapid_commit": schema.BoolAttribute{
				Description: "Enable DHCPv6 rapid commit.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"multi_dns": schema.BoolAttribute{
				Description: "Advertise multiple DNS servers.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"logging": schema.BoolAttribute{
				Description: "Enable DHCP logging.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ignore_unknown_clients": schema.BoolAttribute{
				Description: "Ignore unknown clients (only serve known clients).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *ConfigDHCPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigDHCPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigDHCPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating DHCP config")

	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating DHCP config", err.Error())
		return
	}

	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading DHCP config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDHCPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigDHCPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading DHCP config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDHCPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigDHCPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating DHCP config")

	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating DHCP config", err.Error())
		return
	}

	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading DHCP config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDHCPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Removing DHCP config from state (config remains in Pi-hole)")
}

func (r *ConfigDHCPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing DHCP config from Pi-hole")

	var data ConfigDHCPResourceModel
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error importing DHCP config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDHCPResource) readConfig(ctx context.Context, data *ConfigDHCPResourceModel) error {
	config, err := r.client.GetDHCPConfig(ctx)
	if err != nil {
		return err
	}

	data.ID = types.StringValue("dhcp")
	data.Active = types.BoolValue(config.Active)
	data.Start = types.StringValue(config.Start)
	data.End = types.StringValue(config.End)
	data.Router = types.StringValue(config.Router)
	data.Netmask = types.StringValue(config.Netmask)
	data.LeaseTime = types.StringValue(config.LeaseTime)
	data.IPv6 = types.BoolValue(config.IPv6)
	data.RapidCommit = types.BoolValue(config.RapidCommit)
	data.MultiDNS = types.BoolValue(config.MultiDNS)
	data.Logging = types.BoolValue(config.Logging)
	data.IgnoreUnknownClients = types.BoolValue(config.IgnoreUnknownClients)

	return nil
}

func (r *ConfigDHCPResource) updateConfig(ctx context.Context, data *ConfigDHCPResourceModel) error {
	dhcpConfig := map[string]interface{}{
		"active":               data.Active.ValueBool(),
		"start":                data.Start.ValueString(),
		"end":                  data.End.ValueString(),
		"router":               data.Router.ValueString(),
		"netmask":              data.Netmask.ValueString(),
		"leaseTime":            data.LeaseTime.ValueString(),
		"ipv6":                 data.IPv6.ValueBool(),
		"rapidCommit":          data.RapidCommit.ValueBool(),
		"multiDNS":             data.MultiDNS.ValueBool(),
		"logging":              data.Logging.ValueBool(),
		"ignoreUnknownClients": data.IgnoreUnknownClients.ValueBool(),
	}

	if err := r.client.UpdateConfig(ctx, "dhcp", dhcpConfig); err != nil {
		return fmt.Errorf("failed to update dhcp config: %w", err)
	}

	return nil
}
