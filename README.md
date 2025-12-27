# Terraform Provider for Pi-hole v6

[![Tests](https://github.com/dklesev/terraform-provider-pihole/actions/workflows/test.yml/badge.svg)](https://github.com/dklesev/terraform-provider-pihole/actions/workflows/test.yml)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A Terraform/OpenTofu provider for managing [Pi-hole](https://pi-hole.net/) v6+ instances via the FTL REST API.

> ⚠️ **Important**: This provider requires Pi-hole v6.0 or later. It will NOT work with Pi-hole v5.x.

## Features

- **Full CRUD support** for 18 Pi-hole resources
- **Import support** for all resources
- **Automatic retry logic** for transient network errors
- **Session management** with automatic re-authentication
- Works with both **Terraform** and **OpenTofu**

## Installation

```hcl
terraform {
  required_providers {
    pihole = {
      source  = "dklesev/pihole"
      version = "~> 1.0"
    }
  }
}
```

## Configuration

```hcl
provider "pihole" {
  url      = "http://pi.hole"      # or PIHOLE_URL env var
  password = "your-password"        # or PIHOLE_PASSWORD env var (recommended)
  
  # Optional settings
  timeout                  = 30     # HTTP timeout in seconds
  tls_insecure_skip_verify = false  # Skip TLS certificate verification
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `PIHOLE_URL` | Pi-hole instance URL (e.g., `http://pi.hole`) |
| `PIHOLE_PASSWORD` | Pi-hole web interface password (**recommended** over config) |

> 💡 **Tip**: Use environment variables for the password to avoid storing secrets in state files.

## Quick Start

```hcl
resource "pihole_group" "trusted" {
  name        = "trusted_devices"
  description = "Devices with relaxed blocking"
}

resource "pihole_client" "laptop" {
  client  = "192.168.1.100"
  groups  = [pihole_group.trusted.id]
  comment = "My laptop"
}

resource "pihole_domain" "ads" {
  domain = "ads.example.com"
  type   = "deny"
  kind   = "exact"
}

resource "pihole_list" "blocklist" {
  address = "https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/adblock/pro.txt"
  type    = "block"
}
```

## Resources

### Core Resources

| Resource | Description |
|----------|-------------|
| `pihole_group` | Manage groups for organizing clients and rules |
| `pihole_client` | Manage clients (IP, MAC, hostname, subnet) |
| `pihole_domain` | Manage allow/deny domains (exact/regex) |
| `pihole_list` | Manage blocklist/allowlist subscriptions |

### DNS Resources

| Resource | Description |
|----------|-------------|
| `pihole_dns_blocking` | Control global DNS blocking state |
| `pihole_dns_upstream` | Manage upstream DNS servers |
| `pihole_local_dns` | Manage local A records (hostname → IP) |
| `pihole_cname_record` | Manage local CNAME records |

### DHCP Resources

| Resource | Description |
|----------|-------------|
| `pihole_dhcp_static_lease` | Manage DHCP static leases (MAC → IP) |

### Configuration Resources

| Resource | Description |
|----------|-------------|
| `pihole_config_dns` | DNS server settings (port, caching, rate limiting) |
| `pihole_config_dhcp` | DHCP server settings (range, lease time) |
| `pihole_config_misc` | Miscellaneous settings (privacy, temp units) |
| `pihole_config_ntp` | NTP server configuration |
| `pihole_config_resolver` | Resolver settings (network, refresh) |
| `pihole_config_database` | Database settings (query logging, FTL) |
| `pihole_config_webserver` | Web interface settings (port, session) |
| `pihole_config_files` | File path settings (logs, gravity) |
| `pihole_config_debug` | Debug settings (various debug flags) |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `pihole_groups` | List all groups |
| `pihole_clients` | List all clients |
| `pihole_domains` | List domains (with filtering by type/kind) |
| `pihole_lists` | List subscriptions (with filtering by type) |

## Documentation

Full documentation is available on the [Terraform Registry](https://registry.terraform.io/providers/dklesev/pihole/latest/docs) or in the [`docs/`](./docs) folder.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

MIT - see [LICENSE](LICENSE)
