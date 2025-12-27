# Database Configuration from YAML

resource "pihole_config_database" "settings" {
  db_import       = true
  max_db_days     = 91
  db_interval     = 60
  use_wal         = true
  parse_arp_cache = true
  network_expire  = 91
}
