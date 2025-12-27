# Manage resolver settings for client hostname resolution
resource "pihole_config_resolver" "settings" {
  resolve_ipv4  = true
  resolve_ipv6  = true
  network_names = true
  refresh_names = "IPV4_ONLY"
}
