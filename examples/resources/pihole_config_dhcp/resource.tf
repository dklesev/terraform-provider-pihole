# Manage DHCP server settings
resource "pihole_config_dhcp" "settings" {
  # Enable/disable DHCP
  active = true

  # IP range
  start   = "192.168.1.100"
  end     = "192.168.1.200"
  router  = "192.168.1.1"
  netmask = "255.255.255.0"

  # Lease time
  lease_time = "24h"

  # Optional: IPv6 DHCP
  ipv6         = false
  rapid_commit = false
}
