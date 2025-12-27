# DHCP Configuration from YAML

resource "pihole_config_dhcp" "settings" {
  active                 = local.config.config.dhcp.active
  ipv6                   = local.config.config.dhcp.ipv6
  rapid_commit           = local.config.config.dhcp.rapid_commit
  multi_dns              = local.config.config.dhcp.multi_dns
  logging                = local.config.config.dhcp.logging
  ignore_unknown_clients = local.config.config.dhcp.ignore_unknown_clients
}
