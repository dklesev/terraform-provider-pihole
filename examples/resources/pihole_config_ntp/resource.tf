# Manage NTP time synchronization settings
resource "pihole_config_ntp" "settings" {
  # IPv4 NTP server
  ipv4_active  = true
  ipv4_address = ""

  # IPv6 NTP server
  ipv6_active  = true
  ipv6_address = ""

  # Time sync
  sync_active   = true
  sync_server   = "pool.ntp.org"
  sync_interval = 3600
  sync_count    = 8
}
