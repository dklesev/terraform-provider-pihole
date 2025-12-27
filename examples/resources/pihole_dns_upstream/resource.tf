# Manage DNS upstream servers
resource "pihole_dns_upstream" "google_primary" {
  upstream = "8.8.8.8"
}

resource "pihole_dns_upstream" "google_secondary" {
  upstream = "8.8.4.4"
}

resource "pihole_dns_upstream" "cloudflare" {
  upstream = "1.1.1.1"
}

# Using for_each with a list from YAML config
locals {
  upstreams = ["8.8.8.8", "8.8.4.4", "1.1.1.1"]
}

resource "pihole_dns_upstream" "from_list" {
  for_each = toset(local.upstreams)
  upstream = each.value
}
