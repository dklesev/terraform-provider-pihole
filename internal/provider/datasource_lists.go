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

var _ datasource.DataSource = &ListsDataSource{}

func NewListsDataSource() datasource.DataSource {
	return &ListsDataSource{}
}

type ListsDataSource struct {
	client *client.Client
}

type ListsDataSourceModel struct {
	Type  types.String          `tfsdk:"type"`
	Lists []ListDataSourceModel `tfsdk:"lists"`
}

type ListDataSourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Address   types.String `tfsdk:"address"`
	Type      types.String `tfsdk:"type"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Comment   types.String `tfsdk:"comment"`
	Groups    types.List   `tfsdk:"groups"`
	DateAdded types.Int64  `tfsdk:"date_added"`
	Number    types.Int64  `tfsdk:"number"`
	Status    types.Int64  `tfsdk:"status"`
}

func (d *ListsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lists"
}

func (d *ListsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches Pi-hole list subscriptions with optional filtering.",
		MarkdownDescription: `
Fetches Pi-hole list subscriptions with optional filtering.

## Example Usage

### All Lists

` + "```hcl" + `
data "pihole_lists" "all" {}
` + "```" + `

### Only Blocklists

` + "```hcl" + `
data "pihole_lists" "blocklists" {
  type = "block"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "Filter by type: 'block' or 'allow'. Leave empty for all.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("block", "allow"),
				},
			},
			"lists": schema.ListNestedAttribute{
				Description: "List of list subscriptions matching the filter.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The unique identifier of the list.",
							Computed:    true,
						},
						"address": schema.StringAttribute{
							Description: "The URL of the list.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type: 'block' or 'allow'.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the list is enabled.",
							Computed:    true,
						},
						"comment": schema.StringAttribute{
							Description: "The comment for the list.",
							Computed:    true,
						},
						"groups": schema.ListAttribute{
							Description: "Groups this list applies to.",
							Computed:    true,
							ElementType: types.Int64Type,
						},
						"date_added": schema.Int64Attribute{
							Description: "Unix timestamp when the list was added.",
							Computed:    true,
						},
						"number": schema.Int64Attribute{
							Description: "Number of domains in the list.",
							Computed:    true,
						},
						"status": schema.Int64Attribute{
							Description: "Download status of the list.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ListsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ListsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ListsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listType := ""
	if !data.Type.IsNull() {
		listType = data.Type.ValueString()
	}

	lists, err := d.client.GetLists(ctx, listType, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading lists",
			fmt.Sprintf("Could not read lists: %s", err.Error()),
		)
		return
	}

	data.Lists = make([]ListDataSourceModel, len(lists))
	for i, l := range lists {
		model, diags := mapListToDataSourceModel(ctx, &l)
		resp.Diagnostics.Append(diags...)
		data.Lists[i] = model
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// mapListToDataSourceModel maps a client.List to the data source model.
func mapListToDataSourceModel(ctx context.Context, l *client.List) (ListDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := ListDataSourceModel{
		ID:        types.Int64Value(l.ID),
		Address:   types.StringValue(l.Address),
		Type:      types.StringValue(l.Type),
		Enabled:   types.BoolValue(l.Enabled),
		DateAdded: types.Int64Value(l.DateAdded),
		Number:    types.Int64Value(l.Number),
		Status:    types.Int64Value(int64(l.Status)),
	}

	if l.Comment != "" {
		model.Comment = types.StringValue(l.Comment)
	} else {
		model.Comment = types.StringNull()
	}

	if len(l.Groups) > 0 {
		groupsList, d := types.ListValueFrom(ctx, types.Int64Type, l.Groups)
		diags.Append(d...)
		model.Groups = groupsList
	} else {
		model.Groups = types.ListNull(types.Int64Type)
	}

	return model, diags
}
