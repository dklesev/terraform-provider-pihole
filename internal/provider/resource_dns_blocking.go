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

var _ resource.Resource = &DNSBlockingResource{}

func NewDNSBlockingResource() resource.Resource {
	return &DNSBlockingResource{}
}

type DNSBlockingResource struct {
	client *client.Client
}

type DNSBlockingResourceModel struct {
	ID      types.String  `tfsdk:"id"`
	Enabled types.Bool    `tfsdk:"enabled"`
	Timer   types.Float64 `tfsdk:"timer"`
}

func (r *DNSBlockingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_blocking"
}

func (r *DNSBlockingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the Pi-hole DNS blocking status.",
		MarkdownDescription: `
Manages the Pi-hole DNS blocking status.

This resource controls whether Pi-hole is actively blocking DNS queries.

## Example Usage

### Enable Blocking

` + "```hcl" + `
resource "pihole_dns_blocking" "main" {
  enabled = true
}
` + "```" + `

### Temporarily Disable Blocking

` + "```hcl" + `
resource "pihole_dns_blocking" "main" {
  enabled = false
  timer   = 300  # Auto-enable after 5 minutes
}
` + "```" + `

~> **Note:** This resource is a singleton - only one instance should exist per Pi-hole.
The resource ID is always "blocking".
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether DNS blocking is enabled.",
				Required:    true,
			},
			"timer": schema.Float64Attribute{
				Description: "Seconds until the blocking status automatically toggles. Null for permanent state.",
				Optional:    true,
			},
		},
	}
}

func (r *DNSBlockingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSBlockingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSBlockingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Setting DNS blocking status", map[string]interface{}{
		"enabled": data.Enabled.ValueBool(),
	})

	var timer *float64
	if !data.Timer.IsNull() {
		t := data.Timer.ValueFloat64()
		timer = &t
	}

	result, err := r.client.SetDNSBlocking(ctx, data.Enabled.ValueBool(), timer)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting DNS blocking",
			fmt.Sprintf("Could not set DNS blocking: %s", err.Error()),
		)
		return
	}

	r.mapDNSBlockingToModel(result, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSBlockingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSBlockingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetDNSBlocking(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading DNS blocking",
			fmt.Sprintf("Could not read DNS blocking status: %s", err.Error()),
		)
		return
	}

	data.ID = types.StringValue("blocking")
	data.Enabled = types.BoolValue(result.Blocking == "enabled")
	// Note: We preserve the configured timer value from state rather than reading
	// the countdown value from the API. The API timer counts down in real-time,
	// which would cause constant plan drift. The timer is only meaningful at
	// the time it's set, not as a read value.
	// If the API shows no timer, we set it to null (timer expired or not set).
	if result.Timer == nil {
		data.Timer = types.Float64Null()
	}
	// Otherwise, keep the existing state value (don't update from API)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSBlockingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSBlockingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var timer *float64
	if !data.Timer.IsNull() {
		t := data.Timer.ValueFloat64()
		timer = &t
	}

	result, err := r.client.SetDNSBlocking(ctx, data.Enabled.ValueBool(), timer)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating DNS blocking",
			fmt.Sprintf("Could not update DNS blocking: %s", err.Error()),
		)
		return
	}

	r.mapDNSBlockingToModel(result, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSBlockingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// On delete, we re-enable blocking as the safe default
	tflog.Info(ctx, "Deleting DNS blocking resource - enabling blocking as default")

	_, err := r.client.SetDNSBlocking(ctx, true, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error resetting DNS blocking",
			fmt.Sprintf("Could not reset DNS blocking to enabled: %s", err.Error()),
		)
		return
	}
}

func (r *DNSBlockingResource) mapDNSBlockingToModel(blocking *client.DNSBlocking, data *DNSBlockingResourceModel) {
	data.ID = types.StringValue("blocking")
	data.Enabled = types.BoolValue(blocking.Blocking == "enabled")

	if blocking.Timer != nil {
		data.Timer = types.Float64Value(*blocking.Timer)
	} else {
		data.Timer = types.Float64Null()
	}
}
