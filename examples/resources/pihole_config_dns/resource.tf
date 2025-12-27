# Manage DNS configuration settings
resource "pihole_config_dns" "settings" {
  # DNSSEC validation
  dnssec = true

  # Query logging
  query_logging = true

  # Blocking behavior
  blocking_active = true
  blocking_mode   = "NULL"
  block_ttl       = 2

  # Cache settings
  cache_size      = 10000
  cache_optimizer = 3600

  # Rate limiting
  rate_limit_count    = 1000
  rate_limit_interval = 60

  # Domain settings
  domain_name  = "lan"
  domain_local = true
}
