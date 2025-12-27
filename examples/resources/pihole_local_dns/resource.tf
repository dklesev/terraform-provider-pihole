# Manage local DNS A records
resource "pihole_local_dns" "server" {
  hostname = "server.lan"
  ip       = "192.168.1.100"
}

resource "pihole_local_dns" "nas" {
  hostname = "nas.lan"
  ip       = "192.168.1.50"
}

# Using for_each with a map
locals {
  dns_records = {
    "server.lan" = "192.168.1.100"
    "nas.lan"    = "192.168.1.50"
    "printer.lan" = "192.168.1.30"
  }
}

resource "pihole_local_dns" "records" {
  for_each = local.dns_records
  hostname = each.key
  ip       = each.value
}
