# Manage miscellaneous Pi-hole settings
resource "pihole_config_misc" "settings" {
  # Privacy level (0-3)
  privacy_level = 0

  # System checks
  check_load  = true
  check_shmem = 90
  check_disk  = 90

  # Custom dnsmasq lines
  dnsmasq_lines = [
    "address=/server.lan/192.168.1.100",
    "address=/nas.lan/192.168.1.50"
  ]
}
