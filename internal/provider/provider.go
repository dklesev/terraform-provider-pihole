// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"os"
	"time"

	"github.com/dklesev/terraform-provider-pihole/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure PiholeProvider satisfies various provider interfaces.
var _ provider.Provider = &PiholeProvider{}

// PiholeProvider defines the provider implementation.
type PiholeProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// PiholeProviderModel describes the provider data model.
type PiholeProviderModel struct {
	URL                   types.String `tfsdk:"url"`
	Password              types.String `tfsdk:"password"`
	TLSInsecureSkipVerify types.Bool   `tfsdk:"tls_insecure_skip_verify"`
	Timeout               types.Int64  `tfsdk:"timeout"`
}

func (p *PiholeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pihole"
	resp.Version = p.version
}

func (p *PiholeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Pi-hole v6+ instances via the FTL API.",
		MarkdownDescription: `
The Pi-hole provider allows you to manage Pi-hole instances using Terraform/OpenTofu.

This provider targets **Pi-hole v6.0+** with the new FTL REST API.

## Example Usage

` + "```hcl" + `
provider "pihole" {
  url      = "http://pi.hole"
  password = "your-password"
}
` + "```" + `

## Authentication

The provider supports password-based authentication.

> [!WARNING]
> While you can configure the password in the ` + "`provider`" + ` block, this will result in the password being stored in plain text in the Terraform state file.
> 
> **Strongly Recommended**: Do not set the ` + "`password`" + ` field in the configuration. Instead, set the ` + "`PIHOLE_PASSWORD`" + ` environment variable. This prevents the secret from being persisted in the state.

Configuration options:

1. Environment variables (Recommended): ` + "`PIHOLE_URL`" + `, ` + "`PIHOLE_PASSWORD`" + `
2. Provider configuration block (Not Recommended for secrets)
`,
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "The URL of the Pi-hole instance (e.g., 'http://pi.hole'). Can also be set via the PIHOLE_URL environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password for the Pi-hole web interface. Can also be set via the PIHOLE_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"tls_insecure_skip_verify": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. Default: false.",
				Optional:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "HTTP timeout in seconds. Default: 30.",
				Optional:    true,
			},
		},
	}
}

func (p *PiholeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Pi-hole provider")

	var config PiholeProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use environment variables as fallback
	url := os.Getenv("PIHOLE_URL")
	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}

	password := os.Getenv("PIHOLE_PASSWORD")
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// Validate required configuration
	if url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Missing Pi-hole URL",
			"The provider cannot create the Pi-hole API client without the URL. "+
				"Either set it in the provider configuration or use the PIHOLE_URL environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Build client configuration
	cfg := client.Config{
		URL:      url,
		Password: password,
	}

	if !config.TLSInsecureSkipVerify.IsNull() {
		cfg.TLSInsecureSkipVerify = config.TLSInsecureSkipVerify.ValueBool()
	}

	if !config.Timeout.IsNull() && config.Timeout.ValueInt64() > 0 {
		cfg.Timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	// Create the API client
	apiClient, err := client.New(cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Pi-hole API client",
			"An unexpected error occurred when creating the Pi-hole API client: "+err.Error(),
		)
		return
	}

	// Test authentication
	if err := apiClient.Authenticate(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Failed to authenticate with Pi-hole",
			"The provider was unable to authenticate with the Pi-hole instance: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Pi-hole provider configured successfully", map[string]interface{}{
		"url": url,
	})

	// Make client available to resources and data sources
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

func (p *PiholeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGroupResource,
		NewDomainResource,
		NewClientResource,
		NewListResource,
		NewDNSBlockingResource,
		NewConfigMiscResource,
		NewConfigDNSResource,
		NewConfigDHCPResource,
		NewConfigResolverResource,
		NewConfigDatabaseResource,
		NewConfigDebugResource,
		NewConfigNTPResource,
		NewConfigWebserverResource,
		NewConfigFilesResource,
		NewDNSUpstreamResource,
		NewLocalDNSResource,
		NewCNAMERecordResource,
		NewDHCPStaticLeaseResource,
	}
}

func (p *PiholeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewGroupsDataSource,
		NewDomainsDataSource,
		NewClientsDataSource,
		NewListsDataSource,
	}
}

// New creates a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PiholeProvider{
			version: version,
		}
	}
}
