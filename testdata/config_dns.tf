# DNS Configuration from YAML

resource "pihole_config_dns" "settings" {
  dnssec               = local.config.config.dns.dnssec
  query_logging        = local.config.config.dns.query_logging
  blocking_active      = local.config.config.dns.blocking_active
  cache_size           = local.config.config.dns.cache_size
  cache_optimizer      = local.config.config.dns.cache_optimizer
  rate_limit_count     = local.config.config.dns.rate_limit_count
  rate_limit_interval  = local.config.config.dns.rate_limit_interval
  domain_name          = local.config.config.dns.domain_name
  domain_local         = local.config.config.dns.domain_local
  block_ttl            = local.config.config.dns.block_ttl
  bogus_priv           = local.config.config.dns.bogus_priv
  mozilla_canary       = local.config.config.dns.mozilla_canary
  icloud_private_relay = local.config.config.dns.icloud_private_relay
}
