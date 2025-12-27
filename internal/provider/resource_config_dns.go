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
	_ resource.Resource                = &ConfigDNSResource{}
	_ resource.ResourceWithImportState = &ConfigDNSResource{}
)

func NewConfigDNSResource() resource.Resource {
	return &ConfigDNSResource{}
}

type ConfigDNSResource struct {
	client *client.Client
}

type ConfigDNSResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Port             types.Int64  `tfsdk:"port"`
	Interface        types.String `tfsdk:"interface"`
	ListeningMode    types.String `tfsdk:"listening_mode"`
	DNSSEC           types.Bool   `tfsdk:"dnssec"`
	QueryLogging     types.Bool   `tfsdk:"query_logging"`
	DomainNeeded     types.Bool   `tfsdk:"domain_needed"`
	ExpandHosts      types.Bool   `tfsdk:"expand_hosts"`
	BogusPriv        types.Bool   `tfsdk:"bogus_priv"`
	CNAMEDeepInspect types.Bool   `tfsdk:"cname_deep_inspect"`
	BlockESNI        types.Bool   `tfsdk:"block_esni"`
	BlockTTL         types.Int64  `tfsdk:"block_ttl"`
	PiholePTR        types.String `tfsdk:"pihole_ptr"`
	ReplyWhenBusy    types.String `tfsdk:"reply_when_busy"`
	// Domain settings
	DomainName  types.String `tfsdk:"domain_name"`
	DomainLocal types.Bool   `tfsdk:"domain_local"`
	// Cache settings
	CacheSize      types.Int64 `tfsdk:"cache_size"`
	CacheOptimizer types.Int64 `tfsdk:"cache_optimizer"`
	// Blocking settings
	BlockingActive types.Bool   `tfsdk:"blocking_active"`
	BlockingMode   types.String `tfsdk:"blocking_mode"`
	// Special domains
	MozillaCanary      types.Bool `tfsdk:"mozilla_canary"`
	ICloudPrivateRelay types.Bool `tfsdk:"icloud_private_relay"`
	// Rate limiting
	RateLimitCount    types.Int64 `tfsdk:"rate_limit_count"`
	RateLimitInterval types.Int64 `tfsdk:"rate_limit_interval"`
}

func (r *ConfigDNSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_dns"
}

func (r *ConfigDNSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Pi-hole DNS configuration settings.",
		MarkdownDescription: `
Manages Pi-hole DNS configuration settings including DNSSEC, caching, blocking mode, and rate limiting.

## Example Usage

` + "```hcl" + `
resource "pihole_config_dns" "settings" {
  dnssec        = true
  query_logging = true
  cache_size    = 10000
  
  # Blocking
  blocking_active = true
  blocking_mode   = "NULL"
  
  # Rate limiting
  rate_limit_count    = 1000
  rate_limit_interval = 60
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this resource (always 'dns').",
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "DNS port (default: 53).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(53),
			},
			"interface": schema.StringAttribute{
				Description: "Interface to listen on (empty for all).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"listening_mode": schema.StringAttribute{
				Description: "Listening mode: LOCAL, SINGLE, BIND, ALL.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("LOCAL"),
			},
			"dnssec": schema.BoolAttribute{
				Description: "Enable DNSSEC validation.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"query_logging": schema.BoolAttribute{
				Description: "Enable query logging.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"domain_needed": schema.BoolAttribute{
				Description: "Never forward non-FQDN queries.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"expand_hosts": schema.BoolAttribute{
				Description: "Expand hosts with domain.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"bogus_priv": schema.BoolAttribute{
				Description: "Never forward reverse lookups for private IPs.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"cname_deep_inspect": schema.BoolAttribute{
				Description: "Deep CNAME inspection.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"block_esni": schema.BoolAttribute{
				Description: "Block ESNI/ECH queries.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"block_ttl": schema.Int64Attribute{
				Description: "TTL for blocked queries (seconds).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(2),
			},
			"pihole_ptr": schema.StringAttribute{
				Description: "PTR record for Pi-hole: PI.HOLE, HOSTNAME, HOSTNAMEFQDN, NONE.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("PI.HOLE"),
			},
			"reply_when_busy": schema.StringAttribute{
				Description: "Reply behavior when busy: ALLOW, BLOCK, REFUSE, DROP.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ALLOW"),
			},
			// Domain settings
			"domain_name": schema.StringAttribute{
				Description: "Local domain name.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("lan"),
			},
			"domain_local": schema.BoolAttribute{
				Description: "Domain is local only.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			// Cache settings
			"cache_size": schema.Int64Attribute{
				Description: "DNS cache size.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10000),
			},
			"cache_optimizer": schema.Int64Attribute{
				Description: "Cache optimizer TTL (seconds).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3600),
			},
			// Blocking settings
			"blocking_active": schema.BoolAttribute{
				Description: "Enable blocking.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"blocking_mode": schema.StringAttribute{
				Description: "Blocking mode: NULL, IP-NODATA-AAAA, IP, NXDOMAIN.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("NULL"),
			},
			// Special domains
			"mozilla_canary": schema.BoolAttribute{
				Description: "Block Mozilla's canary domain.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"icloud_private_relay": schema.BoolAttribute{
				Description: "Block iCloud Private Relay.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			// Rate limiting
			"rate_limit_count": schema.Int64Attribute{
				Description: "Rate limit: max queries per interval.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1000),
			},
			"rate_limit_interval": schema.Int64Attribute{
				Description: "Rate limit interval (seconds).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
			},
		},
	}
}

func (r *ConfigDNSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigDNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigDNSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating DNS config")

	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating DNS config", err.Error())
		return
	}

	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading DNS config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigDNSResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading DNS config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigDNSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating DNS config")

	if err := r.updateConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error updating DNS config", err.Error())
		return
	}

	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading DNS config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDNSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Removing DNS config from state (config remains in Pi-hole)")
}

func (r *ConfigDNSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing DNS config from Pi-hole")

	var data ConfigDNSResourceModel
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error importing DNS config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigDNSResource) readConfig(ctx context.Context, data *ConfigDNSResourceModel) error {
	config, err := r.client.GetDNSConfig(ctx)
	if err != nil {
		return err
	}

	data.ID = types.StringValue("dns")
	data.Port = types.Int64Value(int64(config.Port))
	data.Interface = types.StringValue(config.Interface)
	data.ListeningMode = types.StringValue(config.ListeningMode)
	data.DNSSEC = types.BoolValue(config.DNSSEC)
	data.QueryLogging = types.BoolValue(config.QueryLogging)
	data.DomainNeeded = types.BoolValue(config.DomainNeeded)
	data.ExpandHosts = types.BoolValue(config.ExpandHosts)
	data.BogusPriv = types.BoolValue(config.BogusPriv)
	data.CNAMEDeepInspect = types.BoolValue(config.CNAMEDeepInspect)
	data.BlockESNI = types.BoolValue(config.BlockESNI)
	data.BlockTTL = types.Int64Value(int64(config.BlockTTL))
	data.PiholePTR = types.StringValue(config.PiholePTR)
	data.ReplyWhenBusy = types.StringValue(config.ReplyWhenBusy)

	// Domain settings
	if config.Domain != nil {
		data.DomainName = types.StringValue(config.Domain.Name)
		data.DomainLocal = types.BoolValue(config.Domain.Local)
	}

	// Cache settings
	if config.Cache != nil {
		data.CacheSize = types.Int64Value(int64(config.Cache.Size))
		data.CacheOptimizer = types.Int64Value(int64(config.Cache.Optimizer))
	}

	// Blocking settings
	if config.Blocking != nil {
		data.BlockingActive = types.BoolValue(config.Blocking.Active)
		data.BlockingMode = types.StringValue(config.Blocking.Mode)
	}

	// Special domains
	if config.SpecialDomains != nil {
		data.MozillaCanary = types.BoolValue(config.SpecialDomains.MozillaCanary)
		data.ICloudPrivateRelay = types.BoolValue(config.SpecialDomains.ICloudPrivateRelay)
	}

	// Rate limiting
	if config.RateLimit != nil {
		data.RateLimitCount = types.Int64Value(int64(config.RateLimit.Count))
		data.RateLimitInterval = types.Int64Value(int64(config.RateLimit.Interval))
	}

	return nil
}

func (r *ConfigDNSResource) updateConfig(ctx context.Context, data *ConfigDNSResourceModel) error {
	dnsConfig := map[string]interface{}{
		"port":             data.Port.ValueInt64(),
		"interface":        data.Interface.ValueString(),
		"listeningMode":    data.ListeningMode.ValueString(),
		"dnssec":           data.DNSSEC.ValueBool(),
		"queryLogging":     data.QueryLogging.ValueBool(),
		"domainNeeded":     data.DomainNeeded.ValueBool(),
		"expandHosts":      data.ExpandHosts.ValueBool(),
		"bogusPriv":        data.BogusPriv.ValueBool(),
		"CNAMEdeepInspect": data.CNAMEDeepInspect.ValueBool(),
		"blockESNI":        data.BlockESNI.ValueBool(),
		"blockTTL":         data.BlockTTL.ValueInt64(),
		"piholePTR":        data.PiholePTR.ValueString(),
		"replyWhenBusy":    data.ReplyWhenBusy.ValueString(),
		"domain": map[string]interface{}{
			"name":  data.DomainName.ValueString(),
			"local": data.DomainLocal.ValueBool(),
		},
		"cache": map[string]interface{}{
			"size":      data.CacheSize.ValueInt64(),
			"optimizer": data.CacheOptimizer.ValueInt64(),
		},
		"blocking": map[string]interface{}{
			"active": data.BlockingActive.ValueBool(),
			"mode":   data.BlockingMode.ValueString(),
		},
		"specialDomains": map[string]interface{}{
			"mozillaCanary":      data.MozillaCanary.ValueBool(),
			"iCloudPrivateRelay": data.ICloudPrivateRelay.ValueBool(),
		},
		"rateLimit": map[string]interface{}{
			"count":    data.RateLimitCount.ValueInt64(),
			"interval": data.RateLimitInterval.ValueInt64(),
		},
	}

	if err := r.client.UpdateConfig(ctx, "dns", dnsConfig); err != nil {
		return fmt.Errorf("failed to update dns config: %w", err)
	}

	return nil
}
