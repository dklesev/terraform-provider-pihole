# Pi-hole v6 Complete Configuration Example
#
# This example demonstrates managing a complete Pi-hole v6 setup using
# a centralized YAML configuration file and Terraform.

terraform {
  required_providers {
    pihole = {
      source  = "dklesev/pihole"
      version = ">= 0.2.0"
    }
  }
}

provider "pihole" {
  url                     = var.pihole_url
  tls_insecure_skip_verify = true
}

variable "pihole_url" {
  description = "Pi-hole API URL"
  type        = string
  default     = "http://localhost:8080"
}

# Load configuration from YAML file
locals {
  config = yamldecode(file("${path.module}/pihole.yaml"))
}

# ============================================================================
# DNS Configuration
# ============================================================================

resource "pihole_config_dns" "settings" {
  dnssec               = local.config.config.dns.dnssec
  query_logging        = local.config.config.dns.query_logging
  blocking_active      = local.config.config.dns.blocking_active
  cache_size           = local.config.config.dns.cache_size
}

# Upstream DNS servers (from YAML list)
resource "pihole_dns_upstream" "upstreams" {
  for_each = toset(local.config.dns_upstreams)
  upstream = each.value
}

# ============================================================================
# Groups, Clients, Domains, Lists
# ============================================================================

resource "pihole_group" "groups" {
  for_each = local.config.groups
  name     = each.key
  enabled  = each.value.enabled
}

resource "pihole_client" "clients" {
  for_each = local.config.clients
  comment  = each.key
  client   = each.value.ip
  groups   = each.value.groups
}

resource "pihole_list" "blocklists" {
  for_each = toset(local.config.lists.blocklists)
  url      = each.value
  type     = "block"
  enabled  = true
}

# ============================================================================
# Local DNS Records
# ============================================================================

resource "pihole_local_dns" "records" {
  for_each = local.config.local_dns
  hostname = each.key
  ip       = each.value
}
