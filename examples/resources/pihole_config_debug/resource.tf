# Manage debug settings (typically all false in production)
resource "pihole_config_debug" "settings" {
  database   = false
  networking = false
  queries    = false
  api        = false
  resolver   = false
  events     = false
  all        = false
}
