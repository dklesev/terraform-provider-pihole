// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DomainsDataSource{}

func NewDomainsDataSource() datasource.DataSource {
	return &DomainsDataSource{}
}

type DomainsDataSource struct {
	client *client.Client
}

type DomainsDataSourceModel struct {
	Type    types.String            `tfsdk:"type"`
	Kind    types.String            `tfsdk:"kind"`
	Domains []DomainDataSourceModel `tfsdk:"domains"`
}

type DomainDataSourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Domain    types.String `tfsdk:"domain"`
	Type      types.String `tfsdk:"type"`
	Kind      types.String `tfsdk:"kind"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Comment   types.String `tfsdk:"comment"`
	Groups    types.List   `tfsdk:"groups"`
	DateAdded types.Int64  `tfsdk:"date_added"`
}

func (d *DomainsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains"
}

func (d *DomainsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches Pi-hole domain entries with optional filtering.",
		MarkdownDescription: `
Fetches Pi-hole domain entries with optional filtering.

## Example Usage

### All Domains

` + "```hcl" + `
data "pihole_domains" "all" {}
` + "```" + `

### Only Blocked Domains

` + "```hcl" + `
data "pihole_domains" "blocked" {
  type = "deny"
}
` + "```" + `

### Only Regex Rules

` + "```hcl" + `
data "pihole_domains" "regex_rules" {
  kind = "regex"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "Filter by type: 'allow' or 'deny'. Leave empty for all.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("allow", "deny"),
				},
			},
			"kind": schema.StringAttribute{
				Description: "Filter by kind: 'exact' or 'regex'. Leave empty for all.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("exact", "regex"),
				},
			},
			"domains": schema.ListNestedAttribute{
				Description: "List of domains matching the filter.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The unique identifier of the domain.",
							Computed:    true,
						},
						"domain": schema.StringAttribute{
							Description: "The domain name or regex pattern.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type: 'allow' or 'deny'.",
							Computed:    true,
						},
						"kind": schema.StringAttribute{
							Description: "The kind: 'exact' or 'regex'.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the domain entry is enabled.",
							Computed:    true,
						},
						"comment": schema.StringAttribute{
							Description: "The comment for the domain.",
							Computed:    true,
						},
						"groups": schema.ListAttribute{
							Description: "Groups this domain applies to.",
							Computed:    true,
							ElementType: types.Int64Type,
						},
						"date_added": schema.Int64Attribute{
							Description: "Unix timestamp when the domain was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *DomainsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *DomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainType := ""
	if !data.Type.IsNull() {
		domainType = data.Type.ValueString()
	}

	kind := ""
	if !data.Kind.IsNull() {
		kind = data.Kind.ValueString()
	}

	domains, err := d.client.GetDomains(ctx, domainType, kind, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading domains",
			fmt.Sprintf("Could not read domains: %s", err.Error()),
		)
		return
	}

	data.Domains = make([]DomainDataSourceModel, len(domains))
	for i, dom := range domains {
		model, diags := mapDomainToDataSourceModel(ctx, &dom)
		resp.Diagnostics.Append(diags...)
		data.Domains[i] = model
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// mapDomainToDataSourceModel maps a client.Domain to the data source model.
func mapDomainToDataSourceModel(ctx context.Context, dom *client.Domain) (DomainDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := DomainDataSourceModel{
		ID:        types.Int64Value(dom.ID),
		Domain:    types.StringValue(dom.Domain),
		Type:      types.StringValue(dom.Type),
		Kind:      types.StringValue(dom.Kind),
		Enabled:   types.BoolValue(dom.Enabled),
		DateAdded: types.Int64Value(dom.DateAdded),
	}

	if dom.Comment != "" {
		model.Comment = types.StringValue(dom.Comment)
	} else {
		model.Comment = types.StringNull()
	}

	if len(dom.Groups) > 0 {
		groupsList, d := types.ListValueFrom(ctx, types.Int64Type, dom.Groups)
		diags.Append(d...)
		model.Groups = groupsList
	} else {
		model.Groups = types.ListNull(types.Int64Type)
	}

	return model, diags
}
