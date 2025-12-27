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
	_ resource.Resource                = &ConfigResolverResource{}
	_ resource.ResourceWithImportState = &ConfigResolverResource{}
)

func NewConfigResolverResource() resource.Resource {
	return &ConfigResolverResource{}
}

type ConfigResolverResource struct {
	client *client.Client
}

type ConfigResolverResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ResolveIPv4  types.Bool   `tfsdk:"resolve_ipv4"`
	ResolveIPv6  types.Bool   `tfsdk:"resolve_ipv6"`
	NetworkNames types.Bool   `tfsdk:"network_names"`
	RefreshNames types.String `tfsdk:"refresh_names"`
}

func (r *ConfigResolverResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_resolver"
}

func (r *ConfigResolverResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Pi-hole resolver configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"resolve_ipv4": schema.BoolAttribute{
				Description: "Resolve IPv4 addresses.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"resolve_ipv6": schema.BoolAttribute{
				Description: "Resolve IPv6 addresses.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"network_names": schema.BoolAttribute{
				Description: "Resolve network names.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"refresh_names": schema.StringAttribute{
				Description: "Refresh names mode: IPV4_ONLY, IPV4_AND_IPV6, NONE, UNKNOWN.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("IPV4_ONLY"),
			},
		},
	}
}

func (r *ConfigResolverResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigResolverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigResolverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating resolver config", err.Error())
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading resolver config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigResolverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigResolverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading resolver config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigResolverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigResolverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating resolver config", err.Error())
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading resolver config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigResolverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Removing resolver config from state")
}

func (r *ConfigResolverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data ConfigResolverResourceModel
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error importing resolver config", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigResolverResource) readConfig(ctx context.Context, data *ConfigResolverResourceModel) error {
	config, err := r.client.GetResolverConfig(ctx)
	if err != nil {
		return err
	}
	data.ID = types.StringValue("resolver")
	data.ResolveIPv4 = types.BoolValue(config.ResolveIPv4)
	data.ResolveIPv6 = types.BoolValue(config.ResolveIPv6)
	data.NetworkNames = types.BoolValue(config.NetworkNames)
	data.RefreshNames = types.StringValue(config.RefreshNames)
	return nil
}

func (r *ConfigResolverResource) updateConfig(ctx context.Context, data *ConfigResolverResourceModel) error {
	cfg := map[string]interface{}{
		"resolveIPv4":  data.ResolveIPv4.ValueBool(),
		"resolveIPv6":  data.ResolveIPv6.ValueBool(),
		"networkNames": data.NetworkNames.ValueBool(),
		"refreshNames": data.RefreshNames.ValueString(),
	}
	return r.client.UpdateConfig(ctx, "resolver", cfg)
}
