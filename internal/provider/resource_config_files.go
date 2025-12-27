// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ConfigFilesResource{}
	_ resource.ResourceWithImportState = &ConfigFilesResource{}
)

func NewConfigFilesResource() resource.Resource {
	return &ConfigFilesResource{}
}

type ConfigFilesResource struct {
	client *client.Client
}

type ConfigFilesResourceModel struct {
	ID           types.String `tfsdk:"id"`
	PID          types.String `tfsdk:"pid"`
	Database     types.String `tfsdk:"database"`
	Gravity      types.String `tfsdk:"gravity"`
	GravityTmp   types.String `tfsdk:"gravity_tmp"`
	MacVendor    types.String `tfsdk:"mac_vendor"`
	LogFTL       types.String `tfsdk:"log_ftl"`
	LogDnsmasq   types.String `tfsdk:"log_dnsmasq"`
	LogWebserver types.String `tfsdk:"log_webserver"`
}

func (r *ConfigFilesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_files"
}

func (r *ConfigFilesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Pi-hole file paths configuration (read-only for most paths).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"pid": schema.StringAttribute{
				Description: "PID file path.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/run/pihole-FTL.pid"),
			},
			"database": schema.StringAttribute{
				Description: "Database file path.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/etc/pihole/pihole-FTL.db"),
			},
			"gravity": schema.StringAttribute{
				Description: "Gravity database path.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/etc/pihole/gravity.db"),
			},
			"gravity_tmp": schema.StringAttribute{
				Description: "Gravity temp directory.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/tmp"),
			},
			"mac_vendor": schema.StringAttribute{
				Description: "MAC vendor database path.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/macvendor.db"),
			},
			"log_ftl": schema.StringAttribute{
				Description: "FTL log file path.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/var/log/pihole/FTL.log"),
			},
			"log_dnsmasq": schema.StringAttribute{
				Description: "dnsmasq log file path.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/var/log/pihole/pihole.log"),
			},
			"log_webserver": schema.StringAttribute{
				Description: "Webserver log file path.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/var/log/pihole/webserver.log"),
			},
		},
	}
}

func (r *ConfigFilesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigFilesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigFilesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating files config", err.Error())
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading files config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigFilesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigFilesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading files config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigFilesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigFilesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating files config", err.Error())
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading files config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigFilesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Removing files config from state")
}

func (r *ConfigFilesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data ConfigFilesResourceModel
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error importing files config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigFilesResource) readConfig(ctx context.Context, data *ConfigFilesResourceModel) error {
	config, err := r.client.GetFilesConfig(ctx)
	if err != nil {
		return err
	}
	data.ID = types.StringValue("files")
	data.PID = types.StringValue(config.PID)
	data.Database = types.StringValue(config.Database)
	data.Gravity = types.StringValue(config.Gravity)
	data.GravityTmp = types.StringValue(config.GravityTmp)
	data.MacVendor = types.StringValue(config.MACVendor)
	if config.Log != nil {
		data.LogFTL = types.StringValue(config.Log.FTL)
		data.LogDnsmasq = types.StringValue(config.Log.DNSmasq)
		data.LogWebserver = types.StringValue(config.Log.Webserver)
	}
	return nil
}

func (r *ConfigFilesResource) updateConfig(ctx context.Context, data *ConfigFilesResourceModel) error {
	cfg := map[string]interface{}{
		"pid":         data.PID.ValueString(),
		"database":    data.Database.ValueString(),
		"gravity":     data.Gravity.ValueString(),
		"gravity_tmp": data.GravityTmp.ValueString(),
		"macvendor":   data.MacVendor.ValueString(),
		"log": map[string]interface{}{
			"ftl":       data.LogFTL.ValueString(),
			"dnsmasq":   data.LogDnsmasq.ValueString(),
			"webserver": data.LogWebserver.ValueString(),
		},
	}
	return r.client.UpdateConfig(ctx, "files", cfg)
}
