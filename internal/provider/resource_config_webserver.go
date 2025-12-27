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
	_ resource.Resource                = &ConfigWebserverResource{}
	_ resource.ResourceWithImportState = &ConfigWebserverResource{}
)

func NewConfigWebserverResource() resource.Resource {
	return &ConfigWebserverResource{}
}

type ConfigWebserverResource struct {
	client *client.Client
}

type ConfigWebserverResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Domain         types.String `tfsdk:"domain"`
	Port           types.String `tfsdk:"port"`
	Threads        types.Int64  `tfsdk:"threads"`
	ServeAll       types.Bool   `tfsdk:"serve_all"`
	SessionTimeout types.Int64  `tfsdk:"session_timeout"`
	SessionRestore types.Bool   `tfsdk:"session_restore"`
	InterfaceBoxed types.Bool   `tfsdk:"interface_boxed"`
	InterfaceTheme types.String `tfsdk:"interface_theme"`
}

func (r *ConfigWebserverResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_webserver"
}

func (r *ConfigWebserverResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Pi-hole webserver configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"domain": schema.StringAttribute{
				Description: "Webserver domain.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("pi.hole"),
			},
			"port": schema.StringAttribute{
				Description: "Webserver port configuration.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("80o,443os,[::]:80o,[::]:443os"),
			},
			"threads": schema.Int64Attribute{
				Description: "Webserver threads.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(50),
			},
			"serve_all": schema.BoolAttribute{
				Description: "Serve all addresses.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"session_timeout": schema.Int64Attribute{
				Description: "Session timeout in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1800),
			},
			"session_restore": schema.BoolAttribute{
				Description: "Restore sessions on restart.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"interface_boxed": schema.BoolAttribute{
				Description: "Use boxed layout.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"interface_theme": schema.StringAttribute{
				Description: "Interface theme.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("default-auto"),
			},
		},
	}
}

func (r *ConfigWebserverResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigWebserverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigWebserverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating webserver config", err.Error())
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading webserver config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigWebserverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigWebserverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading webserver config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigWebserverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigWebserverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating webserver config", err.Error())
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading webserver config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigWebserverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Removing webserver config from state")
}

func (r *ConfigWebserverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data ConfigWebserverResourceModel
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error importing webserver config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigWebserverResource) readConfig(ctx context.Context, data *ConfigWebserverResourceModel) error {
	config, err := r.client.GetWebserverConfig(ctx)
	if err != nil {
		return err
	}
	data.ID = types.StringValue("webserver")
	data.Domain = types.StringValue(config.Domain)
	data.Port = types.StringValue(config.Port)
	data.Threads = types.Int64Value(int64(config.Threads))
	data.ServeAll = types.BoolValue(config.ServeAll)
	if config.Session != nil {
		data.SessionTimeout = types.Int64Value(int64(config.Session.Timeout))
		data.SessionRestore = types.BoolValue(config.Session.Restore)
	}
	if config.Interface != nil {
		data.InterfaceBoxed = types.BoolValue(config.Interface.Boxed)
		data.InterfaceTheme = types.StringValue(config.Interface.Theme)
	}
	return nil
}

func (r *ConfigWebserverResource) updateConfig(ctx context.Context, data *ConfigWebserverResourceModel) error {
	cfg := map[string]interface{}{
		"domain":    data.Domain.ValueString(),
		"port":      data.Port.ValueString(),
		"threads":   data.Threads.ValueInt64(),
		"serve_all": data.ServeAll.ValueBool(),
		"session": map[string]interface{}{
			"timeout": data.SessionTimeout.ValueInt64(),
			"restore": data.SessionRestore.ValueBool(),
		},
		"interface": map[string]interface{}{
			"boxed": data.InterfaceBoxed.ValueBool(),
			"theme": data.InterfaceTheme.ValueString(),
		},
	}
	return r.client.UpdateConfig(ctx, "webserver", cfg)
}
