// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &GroupsDataSource{}

func NewGroupsDataSource() datasource.DataSource {
	return &GroupsDataSource{}
}

type GroupsDataSource struct {
	client *client.Client
}

type GroupsDataSourceModel struct {
	Groups []GroupDataSourceModel `tfsdk:"groups"`
}

type GroupDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	DateAdded   types.Int64  `tfsdk:"date_added"`
}

func (d *GroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *GroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Pi-hole groups.",
		MarkdownDescription: `
Fetches all Pi-hole groups.

## Example Usage

` + "```hcl" + `
data "pihole_groups" "all" {}

output "group_names" {
  value = [for g in data.pihole_groups.all.groups : g.name]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"groups": schema.ListNestedAttribute{
				Description: "List of all groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The unique identifier of the group.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the group.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the group is enabled.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the group.",
							Computed:    true,
						},
						"date_added": schema.Int64Attribute{
							Description: "Unix timestamp when the group was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *GroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupsDataSourceModel

	groups, err := d.client.GetGroups(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading groups",
			fmt.Sprintf("Could not read groups: %s", err.Error()),
		)
		return
	}

	data.Groups = make([]GroupDataSourceModel, len(groups))
	for i, g := range groups {
		data.Groups[i] = mapGroupToDataSourceModel(&g)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// mapGroupToDataSourceModel maps a client.Group to the data source model.
func mapGroupToDataSourceModel(g *client.Group) GroupDataSourceModel {
	model := GroupDataSourceModel{
		ID:        types.Int64Value(g.ID),
		Name:      types.StringValue(g.Name),
		Enabled:   types.BoolValue(g.Enabled),
		DateAdded: types.Int64Value(g.DateAdded),
	}

	if g.Description != "" {
		model.Description = types.StringValue(g.Description)
	} else {
		model.Description = types.StringNull()
	}

	return model
}
