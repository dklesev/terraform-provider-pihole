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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ConfigNTPResource{}
	_ resource.ResourceWithImportState = &ConfigNTPResource{}
)

func NewConfigNTPResource() resource.Resource {
	return &ConfigNTPResource{}
}

type ConfigNTPResource struct {
	client *client.Client
}

type ConfigNTPResourceModel struct {
	ID           types.String `tfsdk:"id"`
	IPv4Active   types.Bool   `tfsdk:"ipv4_active"`
	IPv4Address  types.String `tfsdk:"ipv4_address"`
	IPv6Active   types.Bool   `tfsdk:"ipv6_active"`
	IPv6Address  types.String `tfsdk:"ipv6_address"`
	SyncActive   types.Bool   `tfsdk:"sync_active"`
	SyncServer   types.String `tfsdk:"sync_server"`
	SyncInterval types.Int64  `tfsdk:"sync_interval"`
	SyncCount    types.Int64  `tfsdk:"sync_count"`
}

func (r *ConfigNTPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_ntp"
}

func (r *ConfigNTPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Pi-hole NTP server configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"ipv4_active": schema.BoolAttribute{
				Description: "Enable IPv4 NTP server.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"ipv4_address": schema.StringAttribute{
				Description: "IPv4 NTP server address.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"ipv6_active": schema.BoolAttribute{
				Description: "Enable IPv6 NTP server.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"ipv6_address": schema.StringAttribute{
				Description: "IPv6 NTP server address.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"sync_active": schema.BoolAttribute{
				Description: "Enable NTP sync.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"sync_server": schema.StringAttribute{
				Description: "NTP sync server.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("pool.ntp.org"),
			},
			"sync_interval": schema.Int64Attribute{
				Description: "NTP sync interval in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3600),
			},
			"sync_count": schema.Int64Attribute{
				Description: "NTP sync count.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(8),
			},
		},
	}
}

func (r *ConfigNTPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigNTPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigNTPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating NTP config", err.Error())
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading NTP config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigNTPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigNTPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading NTP config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigNTPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigNTPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating NTP config", err.Error())
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading NTP config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigNTPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Removing NTP config from state")
}

func (r *ConfigNTPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data ConfigNTPResourceModel
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error importing NTP config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigNTPResource) readConfig(ctx context.Context, data *ConfigNTPResourceModel) error {
	config, err := r.client.GetNTPConfig(ctx)
	if err != nil {
		return err
	}
	data.ID = types.StringValue("ntp")
	if config.IPv4 != nil {
		data.IPv4Active = types.BoolValue(config.IPv4.Active)
		data.IPv4Address = types.StringValue(config.IPv4.Address)
	}
	if config.IPv6 != nil {
		data.IPv6Active = types.BoolValue(config.IPv6.Active)
		data.IPv6Address = types.StringValue(config.IPv6.Address)
	}
	if config.Sync != nil {
		data.SyncActive = types.BoolValue(config.Sync.Active)
		data.SyncServer = types.StringValue(config.Sync.Server)
		data.SyncInterval = types.Int64Value(int64(config.Sync.Interval))
		data.SyncCount = types.Int64Value(int64(config.Sync.Count))
	}
	return nil
}

func (r *ConfigNTPResource) updateConfig(ctx context.Context, data *ConfigNTPResourceModel) error {
	cfg := map[string]interface{}{
		"ipv4": map[string]interface{}{
			"active":  data.IPv4Active.ValueBool(),
			"address": data.IPv4Address.ValueString(),
		},
		"ipv6": map[string]interface{}{
			"active":  data.IPv6Active.ValueBool(),
			"address": data.IPv6Address.ValueString(),
		},
		"sync": map[string]interface{}{
			"active":   data.SyncActive.ValueBool(),
			"server":   data.SyncServer.ValueString(),
			"interval": data.SyncInterval.ValueInt64(),
			"count":    data.SyncCount.ValueInt64(),
		},
	}
	return r.client.UpdateConfig(ctx, "ntp", cfg)
}
