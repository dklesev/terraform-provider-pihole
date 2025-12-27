# Manage database settings for query history
resource "pihole_config_database" "settings" {
  db_import       = true
  max_db_days     = 91      # Keep 3 months of history
  db_interval     = 60      # Write to DB every minute
  use_wal         = true    # Write-Ahead Logging
  parse_arp_cache = true
  network_expire  = 91
}
