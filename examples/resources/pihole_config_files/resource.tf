# Manage file paths (read-mostly, changed only in special cases)
resource "pihole_config_files" "settings" {
  pid          = "/run/pihole-FTL.pid"
  database     = "/etc/pihole/pihole-FTL.db"
  gravity      = "/etc/pihole/gravity.db"
  gravity_tmp  = "/tmp"
  mac_vendor   = "/macvendor.db"
  log_ftl      = "/var/log/pihole/FTL.log"
  log_dnsmasq  = "/var/log/pihole/pihole.log"
  log_webserver = "/var/log/pihole/webserver.log"
}
