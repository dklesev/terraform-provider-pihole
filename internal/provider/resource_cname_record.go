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
	_ resource.Resource                = &CNAMERecordResource{}
	_ resource.ResourceWithImportState = &CNAMERecordResource{}
)

func NewCNAMERecordResource() resource.Resource {
	return &CNAMERecordResource{}
}

type CNAMERecordResource struct {
	client *client.Client
}

type CNAMERecordResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Domain types.String `tfsdk:"domain"`
	Target types.String `tfsdk:"target"`
}

func (r *CNAMERecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cname_record"
}

func (r *CNAMERecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole CNAME record.",
		MarkdownDescription: `
Manages a local CNAME record in Pi-hole.

## Example Usage

` + "```hcl" + `
resource "pihole_cname_record" "www" {
  domain = "www.example.local"
  target = "server.example.local"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier.",
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The domain name (alias).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target": schema.StringAttribute{
				Required:    true,
				Description: "The target domain (canonical name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *CNAMERecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CNAMERecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CNAMERecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Format: "domain,target"
	value := fmt.Sprintf("%s,%s", data.Domain.ValueString(), data.Target.ValueString())
	tflog.Debug(ctx, "Creating CNAME record", map[string]interface{}{"value": value})

	if err := r.client.AddConfigArrayItem(ctx, "dns/cnameRecords", value); err != nil {
		resp.Diagnostics.AddError("Error adding CNAME record", err.Error())
		return
	}

	data.ID = types.StringValue(value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CNAMERecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CNAMERecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	value := fmt.Sprintf("%s,%s", data.Domain.ValueString(), data.Target.ValueString())

	config, err := r.client.GetDNSConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading DNS config", err.Error())
		return
	}

	found := false
	for _, c := range config.CNAMERecords {
		if c == value {
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

func (r *CNAMERecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Changes require replacement")
}

func (r *CNAMERecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CNAMERecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	value := fmt.Sprintf("%s,%s", data.Domain.ValueString(), data.Target.ValueString())
	tflog.Debug(ctx, "Deleting CNAME record", map[string]interface{}{"value": value})

	if err := r.client.DeleteConfigArrayItem(ctx, "dns/cnameRecords", value); err != nil {
		resp.Diagnostics.AddError("Error deleting CNAME record", err.Error())
		return
	}
}

func (r *CNAMERecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "domain,target"
	parts := strings.SplitN(req.ID, ",", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: 'domain,target'")
		return
	}

	data := CNAMERecordResourceModel{
		ID:     types.StringValue(req.ID),
		Domain: types.StringValue(parts[0]),
		Target: types.StringValue(parts[1]),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
