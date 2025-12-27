# Block a specific domain (exact match)
resource "pihole_domain" "block_ads" {
  domain  = "ads.example.com"
  type    = "deny"
  kind    = "exact"
  enabled = true
  comment = "Block ads domain"
}

# Allow a domain using regex pattern
resource "pihole_domain" "allow_google" {
  domain  = "^.*\\.google\\.com$"
  type    = "allow"
  kind    = "regex"
  enabled = true
  comment = "Allow all Google subdomains"
}

# Block domains matching a regex pattern, applied to specific group
resource "pihole_domain" "block_trackers" {
  domain  = "^tracking\\..*"
  type    = "deny"
  kind    = "regex"
  enabled = true
  groups  = [pihole_group.iot.id]
  comment = "Block tracking domains for IoT devices"
}
