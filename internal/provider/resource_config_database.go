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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ConfigDatabaseResource{}
	_ resource.ResourceWithImportState = &ConfigDatabaseResource{}
)

func NewConfigDatabaseResource() resource.Resource {
	return &ConfigDatabaseResource{}
}

type ConfigDatabaseResource struct {
	client *client.Client
}

type ConfigDatabaseResourceModel struct {
	ID            types.String `tfsdk:"id"`
	DBImport      types.Bool   `tfsdk:"db_import"`
	MaxDBDays     types.Int64  `tfsdk:"max_db_days"`
	DBInterval    types.Int64  `tfsdk:"db_interval"`
	UseWAL        types.Bool   `tfsdk:"use_wal"`
	ParseARPCache types.Bool   `tfsdk:"parse_arp_cache"`
	NetworkExpire types.Int64  `tfsdk:"network_expire"`
}

func (r *ConfigDatabaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_database"
}

func (r *ConfigDatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Pi-hole database configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"db_import": schema.BoolAttribute{
				Description: "Import database on startup.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"max_db_days": schema.Int64Attribute{
				Description: "Maximum database history in days.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(91),
			},
			"db_interval": schema.Int64Attribute{
				Description: "Database write interval in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
			},
			"use_wal": schema.BoolAttribute{
				Description: "Use WAL mode for database.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"parse_arp_cache": schema.BoolAttribute{
				Description: "Parse ARP cache for network table.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"network_expire": schema.Int64Attribute{
				Description: "Network table entry expiration in days.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(91),
			},
		},
	}
}

func (r *ConfigDatabaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigDatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigDatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating database config", err.Error())
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading database config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigDatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading database config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigDatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating database config", err.Error())
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading database config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Removing database config from state")
}

func (r *ConfigDatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data ConfigDatabaseResourceModel
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error importing database config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDatabaseResource) readConfig(ctx context.Context, data *ConfigDatabaseResourceModel) error {
	config, err := r.client.GetDatabaseConfig(ctx)
	if err != nil {
		return err
	}
	data.ID = types.StringValue("database")
	data.DBImport = types.BoolValue(config.DBImport)
	data.MaxDBDays = types.Int64Value(int64(config.MaxDBDays))
	data.DBInterval = types.Int64Value(int64(config.DBInterval))
	data.UseWAL = types.BoolValue(config.UseWAL)
	if config.Network != nil {
		data.ParseARPCache = types.BoolValue(config.Network.ParseARPCache)
		data.NetworkExpire = types.Int64Value(int64(config.Network.Expire))
	}
	return nil
}

func (r *ConfigDatabaseResource) updateConfig(ctx context.Context, data *ConfigDatabaseResourceModel) error {
	cfg := map[string]interface{}{
		"DBimport":   data.DBImport.ValueBool(),
		"maxDBdays":  data.MaxDBDays.ValueInt64(),
		"DBinterval": data.DBInterval.ValueInt64(),
		"useWAL":     data.UseWAL.ValueBool(),
		"network": map[string]interface{}{
			"parseARPcache": data.ParseARPCache.ValueBool(),
			"expire":        data.NetworkExpire.ValueInt64(),
		},
	}
	return r.client.UpdateConfig(ctx, "database", cfg)
}
