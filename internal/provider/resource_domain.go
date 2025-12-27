// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &DomainResource{}
	_ resource.ResourceWithImportState = &DomainResource{}
)

func NewDomainResource() resource.Resource {
	return &DomainResource{}
}

type DomainResource struct {
	client *client.Client
}

type DomainResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Domain       types.String `tfsdk:"domain"`
	Type         types.String `tfsdk:"type"`
	Kind         types.String `tfsdk:"kind"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Comment      types.String `tfsdk:"comment"`
	Groups       types.Set    `tfsdk:"groups"`
	DateAdded    types.Int64  `tfsdk:"date_added"`
	DateModified types.Int64  `tfsdk:"date_modified"`
}

func (r *DomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *DomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole domain entry for allow/deny lists.",
		MarkdownDescription: `
Manages a Pi-hole domain entry for allow/deny lists.

Domains can be added to either the allow list (whitelist) or deny list (blacklist),
and can be either exact matches or regular expressions.

## Example Usage

### Exact Domain Block

` + "```hcl" + `
resource "pihole_domain" "block_ads" {
  domain  = "ads.example.com"
  type    = "deny"
  kind    = "exact"
  enabled = true
  comment = "Block ads domain"
}
` + "```" + `

### Regex Allow Rule

` + "```hcl" + `
resource "pihole_domain" "allow_google" {
  domain  = "^.*\\.google\\.com$"
  type    = "allow"
  kind    = "regex"
  enabled = true
}
` + "```" + `

## Import

Domains can be imported using the format ` + "`type/kind/domain`" + `:

` + "```shell" + `
terraform import pihole_domain.example deny/exact/ads.example.com
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the domain in Pi-hole.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "The domain name or regex pattern.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of domain entry: 'allow' or 'deny'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("allow", "deny"),
				},
			},
			"kind": schema.StringAttribute{
				Description: "The kind of domain entry: 'exact' or 'regex'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("exact", "regex"),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the domain entry is enabled. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"comment": schema.StringAttribute{
				Description: "A comment describing the domain entry.",
				Optional:    true,
			},
			"groups": schema.SetAttribute{
				Description: "List of group IDs this domain applies to. Default group ID is 0.",
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"date_added": schema.Int64Attribute{
				Description: "Unix timestamp when the domain was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"date_modified": schema.Int64Attribute{
				Description: "Unix timestamp when the domain was last modified.",
				Computed:    true,
			},
		},
	}
}

func (r *DomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating domain", map[string]interface{}{
		"domain": data.Domain.ValueString(),
		"type":   data.Type.ValueString(),
		"kind":   data.Kind.ValueString(),
	})

	var groups []int64
	if !data.Groups.IsNull() && !data.Groups.IsUnknown() {
		resp.Diagnostics.Append(data.Groups.ElementsAs(ctx, &groups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	domain := &client.Domain{
		Domain:  data.Domain.ValueString(),
		Type:    data.Type.ValueString(),
		Kind:    data.Kind.ValueString(),
		Enabled: data.Enabled.ValueBool(),
		Comment: data.Comment.ValueString(),
		Groups:  groups,
	}

	created, err := r.client.CreateDomain(ctx, domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating domain",
			fmt.Sprintf("Could not create domain %s: %s", data.Domain.ValueString(), err.Error()),
		)
		return
	}

	r.mapDomainToModel(ctx, created, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.GetDomain(ctx, data.Type.ValueString(), data.Kind.ValueString(), data.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading domain",
			fmt.Sprintf("Could not read domain %s: %s", data.Domain.ValueString(), err.Error()),
		)
		return
	}

	if domain == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapDomainToModel(ctx, domain, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainResourceModel
	var state DomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var groups []int64
	if !data.Groups.IsNull() && !data.Groups.IsUnknown() {
		resp.Diagnostics.Append(data.Groups.ElementsAs(ctx, &groups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	domain := &client.Domain{
		Domain:  data.Domain.ValueString(),
		Type:    data.Type.ValueString(),
		Kind:    data.Kind.ValueString(),
		Enabled: data.Enabled.ValueBool(),
		Comment: data.Comment.ValueString(),
		Groups:  groups,
	}

	updated, err := r.client.UpdateDomain(ctx,
		state.Type.ValueString(),
		state.Kind.ValueString(),
		state.Domain.ValueString(),
		domain,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating domain",
			fmt.Sprintf("Could not update domain %s: %s", state.Domain.ValueString(), err.Error()),
		)
		return
	}

	r.mapDomainToModel(ctx, updated, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDomain(ctx, data.Type.ValueString(), data.Kind.ValueString(), data.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting domain",
			fmt.Sprintf("Could not delete domain %s: %s", data.Domain.ValueString(), err.Error()),
		)
		return
	}
}

func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: type/kind/domain
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format: type/kind/domain (e.g., deny/exact/ads.example.com)",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("kind"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[2])...)
}

func (r *DomainResource) mapDomainToModel(ctx context.Context, domain *client.Domain, data *DomainResourceModel, diags *diag.Diagnostics) {
	data.ID = types.Int64Value(domain.ID)
	data.Domain = types.StringValue(domain.Domain)
	data.Type = types.StringValue(domain.Type)
	data.Kind = types.StringValue(domain.Kind)
	data.Enabled = types.BoolValue(domain.Enabled)

	if domain.Comment != "" {
		data.Comment = types.StringValue(domain.Comment)
	} else {
		data.Comment = types.StringNull()
	}

	if len(domain.Groups) > 0 {
		groupsList, d := types.SetValueFrom(ctx, types.Int64Type, domain.Groups)
		diags.Append(d...)
		data.Groups = groupsList
	} else {
		data.Groups = types.SetNull(types.Int64Type)
	}

	data.DateAdded = types.Int64Value(domain.DateAdded)
	data.DateModified = types.Int64Value(domain.DateModified)
}
