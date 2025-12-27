# Manage DHCP static leases (MAC to IP reservations)
resource "pihole_dhcp_static_lease" "server" {
  mac      = "AA:BB:CC:DD:EE:FF"
  ip       = "192.168.1.100"
  hostname = "server"
}

resource "pihole_dhcp_static_lease" "nas" {
  mac      = "11:22:33:44:55:66"
  ip       = "192.168.1.50"
  hostname = "nas"
}
