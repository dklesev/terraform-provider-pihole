# Misc Configuration from YAML

resource "pihole_config_misc" "settings" {
  privacy_level = local.config.config.misc.privacy_level
  delay_startup = local.config.config.misc.delay_startup
  nice          = local.config.config.misc.nice
  etc_dnsmasq_d = local.config.config.misc.etc_dnsmasq_d
  extra_logging = local.config.config.misc.extra_logging
  read_only     = local.config.config.misc.read_only
  check_load    = local.config.config.misc.check_load
  check_shmem   = local.config.config.misc.check_shmem
  check_disk    = local.config.config.misc.check_disk
  dnsmasq_lines = local.config.config.misc.dnsmasq_lines
}
