# Enable DNS blocking (default state)
resource "pihole_dns_blocking" "main" {
  enabled = true
}

# Temporarily disable DNS blocking for 5 minutes
resource "pihole_dns_blocking" "temporary_disable" {
  enabled = false
  timer   = 300 # seconds
}
