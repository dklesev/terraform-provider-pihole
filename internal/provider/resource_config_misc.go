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
	_ resource.Resource                = &ConfigMiscResource{}
	_ resource.ResourceWithImportState = &ConfigMiscResource{}
)

func NewConfigMiscResource() resource.Resource {
	return &ConfigMiscResource{}
}

type ConfigMiscResource struct {
	client *client.Client
}

type ConfigMiscResourceModel struct {
	ID              types.String `tfsdk:"id"`
	PrivacyLevel    types.Int64  `tfsdk:"privacy_level"`
	DelayStartup    types.Int64  `tfsdk:"delay_startup"`
	Nice            types.Int64  `tfsdk:"nice"`
	Addr2Line       types.Bool   `tfsdk:"addr2line"`
	EtcDnsmasqD     types.Bool   `tfsdk:"etc_dnsmasq_d"`
	DnsmasqLines    types.List   `tfsdk:"dnsmasq_lines"`
	ExtraLogging    types.Bool   `tfsdk:"extra_logging"`
	ReadOnly        types.Bool   `tfsdk:"read_only"`
	NormalizeCPU    types.Bool   `tfsdk:"normalize_cpu"`
	HideDnsmasqWarn types.Bool   `tfsdk:"hide_dnsmasq_warn"`
	CheckLoad       types.Bool   `tfsdk:"check_load"`
	CheckShmem      types.Int64  `tfsdk:"check_shmem"`
	CheckDisk       types.Int64  `tfsdk:"check_disk"`
}

func (r *ConfigMiscResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_misc"
}

func (r *ConfigMiscResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Pi-hole miscellaneous configuration settings.",
		MarkdownDescription: `
Manages Pi-hole miscellaneous configuration settings including custom dnsmasq lines.

## Example Usage

` + "```hcl" + `
resource "pihole_config_misc" "settings" {
  privacy_level = 0
  etc_dnsmasq_d = false
  
  dnsmasq_lines = [
    "address=/custom.local/192.168.1.100",
    "server=/corp.local/10.0.0.1"
  ]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this resource (always 'misc').",
				Computed:    true,
			},
			"privacy_level": schema.Int64Attribute{
				Description: "Privacy level for statistics (0-3). 0=show everything, 3=hide everything.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"delay_startup": schema.Int64Attribute{
				Description: "Delay FTL startup by this many seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"nice": schema.Int64Attribute{
				Description: "Process priority (nice value).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(-10),
			},
			"addr2line": schema.BoolAttribute{
				Description: "Enable stack trace support for debugging.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"etc_dnsmasq_d": schema.BoolAttribute{
				Description: "Load configuration files from /etc/dnsmasq.d.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"dnsmasq_lines": schema.ListAttribute{
				Description: "Custom dnsmasq configuration lines.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"extra_logging": schema.BoolAttribute{
				Description: "Enable extra debug logging.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"read_only": schema.BoolAttribute{
				Description: "Enable read-only mode (no configuration changes allowed).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"normalize_cpu": schema.BoolAttribute{
				Description: "Normalize CPU load across all cores.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"hide_dnsmasq_warn": schema.BoolAttribute{
				Description: "Hide dnsmasq warnings in the log.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"check_load": schema.BoolAttribute{
				Description: "Enable system load checking.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"check_shmem": schema.Int64Attribute{
				Description: "Shared memory usage threshold (%).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(90),
			},
			"check_disk": schema.Int64Attribute{
				Description: "Disk usage threshold (%).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(90),
			},
		},
	}
}

func (r *ConfigMiscResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigMiscResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigMiscResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating misc config")

	// Build the config update
	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating misc config", err.Error())
		return
	}

	// Read back the config to get computed values
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading misc config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigMiscResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigMiscResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading misc config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigMiscResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigMiscResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating misc config")

	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating misc config", err.Error())
		return
	}

	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading misc config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigMiscResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Config resources don't really "delete" - we just remove from state
	tflog.Debug(ctx, "Removing misc config from state (config remains in Pi-hole)")
}

func (r *ConfigMiscResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// For config resources, the import ID is ignored - we just read the current config
	tflog.Debug(ctx, "Importing misc config from Pi-hole")

	var data ConfigMiscResourceModel
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error importing misc config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigMiscResource) readConfig(ctx context.Context, data *ConfigMiscResourceModel) error {
	config, err := r.client.GetMiscConfig(ctx)
	if err != nil {
		return err
	}

	// Set ID for singleton resource
	data.ID = types.StringValue("misc")

	data.PrivacyLevel = types.Int64Value(int64(config.PrivacyLevel))
	data.DelayStartup = types.Int64Value(int64(config.DelayStartup))
	data.Nice = types.Int64Value(int64(config.Nice))
	data.Addr2Line = types.BoolValue(config.Addr2Line)
	data.EtcDnsmasqD = types.BoolValue(config.EtcDnsmasqD)
	data.ExtraLogging = types.BoolValue(config.ExtraLogging)
	data.ReadOnly = types.BoolValue(config.ReadOnly)
	data.NormalizeCPU = types.BoolValue(config.NormalizeCPU)
	data.HideDnsmasqWarn = types.BoolValue(config.HideDnsmasqWarn)

	// Handle dnsmasq_lines - always use a list value (empty or populated)
	// to maintain consistency with Terraform state
	lines, diags := types.ListValueFrom(ctx, types.StringType, config.DnsmasqLines)
	if diags.HasError() {
		return fmt.Errorf("failed to convert dnsmasq_lines")
	}
	data.DnsmasqLines = lines

	// Handle check settings
	if config.Check != nil {
		data.CheckLoad = types.BoolValue(config.Check.Load)
		data.CheckShmem = types.Int64Value(int64(config.Check.Shmem))
		data.CheckDisk = types.Int64Value(int64(config.Check.Disk))
	}

	return nil
}

func (r *ConfigMiscResource) updateConfig(ctx context.Context, data *ConfigMiscResourceModel) error {
	// Build the misc config values
	miscConfig := map[string]interface{}{
		"privacylevel":      data.PrivacyLevel.ValueInt64(),
		"delay_startup":     data.DelayStartup.ValueInt64(),
		"nice":              data.Nice.ValueInt64(),
		"addr2line":         data.Addr2Line.ValueBool(),
		"etc_dnsmasq_d":     data.EtcDnsmasqD.ValueBool(),
		"extraLogging":      data.ExtraLogging.ValueBool(),
		"readOnly":          data.ReadOnly.ValueBool(),
		"normalizeCPU":      data.NormalizeCPU.ValueBool(),
		"hide_dnsmasq_warn": data.HideDnsmasqWarn.ValueBool(),
		"check": map[string]interface{}{
			"load":  data.CheckLoad.ValueBool(),
			"shmem": data.CheckShmem.ValueInt64(),
			"disk":  data.CheckDisk.ValueInt64(),
		},
	}

	// Handle dnsmasq_lines
	if !data.DnsmasqLines.IsNull() && !data.DnsmasqLines.IsUnknown() {
		var lines []string
		if diags := data.DnsmasqLines.ElementsAs(ctx, &lines, false); diags.HasError() {
			return fmt.Errorf("failed to parse dnsmasq_lines")
		}
		miscConfig["dnsmasq_lines"] = lines
	}

	// Send single PATCH request with all misc config
	if err := r.client.UpdateConfig(ctx, "misc", miscConfig); err != nil {
		return fmt.Errorf("failed to update misc config: %w", err)
	}

	return nil
}
