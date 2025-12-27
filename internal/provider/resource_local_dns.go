// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &LocalDNSResource{}
	_ resource.ResourceWithImportState = &LocalDNSResource{}
)

func NewLocalDNSResource() resource.Resource {
	return &LocalDNSResource{}
}

type LocalDNSResource struct {
	client *client.Client
}

type LocalDNSResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Hostname types.String `tfsdk:"hostname"`
	IP       types.String `tfsdk:"ip"`
}

func (r *LocalDNSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_local_dns"
}

func (r *LocalDNSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole local DNS record (A record).",
		MarkdownDescription: `
Manages a local DNS A record in Pi-hole (hostname -> IP mapping).

## Example Usage

` + "```hcl" + `
resource "pihole_local_dns" "server" {
  hostname = "server.lan"
  ip       = "192.168.1.100"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier (hostname IP).",
			},
			"hostname": schema.StringAttribute{
				Required:    true,
				Description: "The hostname for the DNS record.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip": schema.StringAttribute{
				Required:    true,
				Description: "The IP address for the DNS record.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *LocalDNSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LocalDNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LocalDNSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Format: "IP hostname" - hosts file format
	value := fmt.Sprintf("%s %s", data.IP.ValueString(), data.Hostname.ValueString())
	tflog.Debug(ctx, "Creating local DNS", map[string]interface{}{"value": value})

	if err := r.client.AddConfigArrayItem(ctx, "dns/hosts", value); err != nil {
		resp.Diagnostics.AddError("Error adding local DNS", err.Error())
		return
	}

	data.ID = types.StringValue(value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LocalDNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LocalDNSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	value := fmt.Sprintf("%s %s", data.IP.ValueString(), data.Hostname.ValueString())

	config, err := r.client.GetDNSConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading DNS config", err.Error())
		return
	}

	found := false
	for _, h := range config.Hosts {
		if h == value {
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LocalDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Changes require replacement")
}

func (r *LocalDNSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LocalDNSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	value := fmt.Sprintf("%s %s", data.IP.ValueString(), data.Hostname.ValueString())
	tflog.Debug(ctx, "Deleting local DNS", map[string]interface{}{"value": value})

	if err := r.client.DeleteConfigArrayItem(ctx, "dns/hosts", value); err != nil {
		resp.Diagnostics.AddError("Error deleting local DNS", err.Error())
		return
	}
}

func (r *LocalDNSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "IP hostname"
	parts := strings.SplitN(req.ID, " ", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: 'IP hostname'")
		return
	}

	data := LocalDNSResourceModel{
		ID:       types.StringValue(req.ID),
		IP:       types.StringValue(parts[0]),
		Hostname: types.StringValue(parts[1]),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
