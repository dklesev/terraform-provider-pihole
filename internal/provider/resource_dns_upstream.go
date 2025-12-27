// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &DNSUpstreamResource{}
	_ resource.ResourceWithImportState = &DNSUpstreamResource{}
)

func NewDNSUpstreamResource() resource.Resource {
	return &DNSUpstreamResource{}
}

type DNSUpstreamResource struct {
	client *client.Client
}

type DNSUpstreamResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Upstream types.String `tfsdk:"upstream"`
}

func (r *DNSUpstreamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_upstream"
}

func (r *DNSUpstreamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole DNS upstream server.",
		MarkdownDescription: `
Manages a single DNS upstream server in Pi-hole. Each upstream is an individual resource.

## Example Usage

` + "```hcl" + `
resource "pihole_dns_upstream" "google_primary" {
  upstream = "8.8.8.8"
}

resource "pihole_dns_upstream" "google_secondary" {
  upstream = "8.8.4.4"
}

resource "pihole_dns_upstream" "cloudflare" {
  upstream = "1.1.1.1"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier (same as upstream).",
			},
			"upstream": schema.StringAttribute{
				Required:    true,
				Description: "Upstream DNS server address (IP or hostname, optionally with port).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *DNSUpstreamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSUpstreamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSUpstreamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	upstream := data.Upstream.ValueString()
	tflog.Debug(ctx, "Creating DNS upstream", map[string]interface{}{"upstream": upstream})

	// PUT /api/config/dns/upstreams/{upstream}
	if err := r.client.AddConfigArrayItem(ctx, "dns/upstreams", upstream); err != nil {
		resp.Diagnostics.AddError("Error adding DNS upstream", err.Error())
		return
	}

	data.ID = types.StringValue(upstream)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSUpstreamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSUpstreamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	upstream := data.Upstream.ValueString()

	// Check if upstream still exists
	config, err := r.client.GetDNSConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading DNS config", err.Error())
		return
	}

	found := false
	for _, u := range config.Upstreams {
		if u == upstream {
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(upstream)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSUpstreamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Upstream changes require replace, so Update should not be called
	resp.Diagnostics.AddError("Update not supported", "Upstream changes require replacement")
}

func (r *DNSUpstreamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSUpstreamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	upstream := data.Upstream.ValueString()
	tflog.Debug(ctx, "Deleting DNS upstream", map[string]interface{}{"upstream": upstream})

	// DELETE /api/config/dns/upstreams/{upstream}
	if err := r.client.DeleteConfigArrayItem(ctx, "dns/upstreams", upstream); err != nil {
		resp.Diagnostics.AddError("Error deleting DNS upstream", err.Error())
		return
	}
}

func (r *DNSUpstreamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	upstream := req.ID

	// Verify it exists
	config, err := r.client.GetDNSConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading DNS config", err.Error())
		return
	}

	found := false
	for _, u := range config.Upstreams {
		if u == upstream {
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("Upstream not found", fmt.Sprintf("Upstream %q not found in Pi-hole", upstream))
		return
	}

	data := DNSUpstreamResourceModel{
		ID:       types.StringValue(upstream),
		Upstream: types.StringValue(upstream),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
