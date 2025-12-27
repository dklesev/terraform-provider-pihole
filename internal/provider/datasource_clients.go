// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ClientsDataSource{}

func NewClientsDataSource() datasource.DataSource {
	return &ClientsDataSource{}
}

type ClientsDataSource struct {
	client *client.Client
}

type ClientsDataSourceModel struct {
	Clients []ClientDataSourceModel `tfsdk:"clients"`
}

type ClientDataSourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Client    types.String `tfsdk:"client"`
	Comment   types.String `tfsdk:"comment"`
	Groups    types.List   `tfsdk:"groups"`
	DateAdded types.Int64  `tfsdk:"date_added"`
}

func (d *ClientsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clients"
}

func (d *ClientsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Pi-hole client configurations.",
		MarkdownDescription: `
Fetches all Pi-hole client configurations.

## Example Usage

` + "```hcl" + `
data "pihole_clients" "all" {}

output "client_count" {
  value = length(data.pihole_clients.all.clients)
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"clients": schema.ListNestedAttribute{
				Description: "List of all client configurations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The unique identifier of the client.",
							Computed:    true,
						},
						"client": schema.StringAttribute{
							Description: "The client identifier.",
							Computed:    true,
						},
						"comment": schema.StringAttribute{
							Description: "The comment for the client.",
							Computed:    true,
						},
						"groups": schema.ListAttribute{
							Description: "Groups this client belongs to.",
							Computed:    true,
							ElementType: types.Int64Type,
						},
						"date_added": schema.Int64Attribute{
							Description: "Unix timestamp when the client was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ClientsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClientsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClientsDataSourceModel

	clients, err := d.client.GetClients(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading clients",
			fmt.Sprintf("Could not read clients: %s", err.Error()),
		)
		return
	}

	data.Clients = make([]ClientDataSourceModel, len(clients))
	for i, c := range clients {
		model, diags := mapClientToDataSourceModel(ctx, &c)
		resp.Diagnostics.Append(diags...)
		data.Clients[i] = model
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// mapClientToDataSourceModel maps a client.PiholeClient to the data source model.
func mapClientToDataSourceModel(ctx context.Context, c *client.PiholeClient) (ClientDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := ClientDataSourceModel{
		ID:        types.Int64Value(c.ID),
		Client:    types.StringValue(c.Client),
		DateAdded: types.Int64Value(c.DateAdded),
	}

	if c.Comment != "" {
		model.Comment = types.StringValue(c.Comment)
	} else {
		model.Comment = types.StringNull()
	}

	if len(c.Groups) > 0 {
		groupsList, d := types.ListValueFrom(ctx, types.Int64Type, c.Groups)
		diags.Append(d...)
		model.Groups = groupsList
	} else {
		model.Groups = types.ListNull(types.Int64Type)
	}

	return model, diags
}
