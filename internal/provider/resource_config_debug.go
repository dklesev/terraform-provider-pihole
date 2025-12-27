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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ConfigDebugResource{}
	_ resource.ResourceWithImportState = &ConfigDebugResource{}
)

func NewConfigDebugResource() resource.Resource {
	return &ConfigDebugResource{}
}

type ConfigDebugResource struct {
	client *client.Client
}

type ConfigDebugResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Database   types.Bool   `tfsdk:"database"`
	Networking types.Bool   `tfsdk:"networking"`
	Queries    types.Bool   `tfsdk:"queries"`
	API        types.Bool   `tfsdk:"api"`
	Resolver   types.Bool   `tfsdk:"resolver"`
	Events     types.Bool   `tfsdk:"events"`
	All        types.Bool   `tfsdk:"all"`
}

func (r *ConfigDebugResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_debug"
}

func (r *ConfigDebugResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Pi-hole debug configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"database": schema.BoolAttribute{
				Description: "Enable database debugging.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"networking": schema.BoolAttribute{
				Description: "Enable networking debugging.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"queries": schema.BoolAttribute{
				Description: "Enable query debugging.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"api": schema.BoolAttribute{
				Description: "Enable API debugging.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"resolver": schema.BoolAttribute{
				Description: "Enable resolver debugging.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"events": schema.BoolAttribute{
				Description: "Enable events debugging.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"all": schema.BoolAttribute{
				Description: "Enable all debugging (overrides individual settings).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *ConfigDebugResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigDebugResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigDebugResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating debug config", err.Error())
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading debug config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDebugResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigDebugResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading debug config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDebugResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigDebugResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating debug config", err.Error())
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading debug config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDebugResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Removing debug config from state")
}

func (r *ConfigDebugResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data ConfigDebugResourceModel
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error importing debug config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDebugResource) readConfig(ctx context.Context, data *ConfigDebugResourceModel) error {
	config, err := r.client.GetDebugConfig(ctx)
	if err != nil {
		return err
	}
	data.ID = types.StringValue("debug")
	data.Database = types.BoolValue(config.Database)
	data.Networking = types.BoolValue(config.Networking)
	data.Queries = types.BoolValue(config.Queries)
	data.API = types.BoolValue(config.API)
	data.Resolver = types.BoolValue(config.Resolver)
	data.Events = types.BoolValue(config.Events)
	data.All = types.BoolValue(config.All)
	return nil
}

func (r *ConfigDebugResource) updateConfig(ctx context.Context, data *ConfigDebugResourceModel) error {
	cfg := map[string]interface{}{
		"database":   data.Database.ValueBool(),
		"networking": data.Networking.ValueBool(),
		"queries":    data.Queries.ValueBool(),
		"api":        data.API.ValueBool(),
		"resolver":   data.Resolver.ValueBool(),
		"events":     data.Events.ValueBool(),
		"all":        data.All.ValueBool(),
	}
	return r.client.UpdateConfig(ctx, "debug", cfg)
}
