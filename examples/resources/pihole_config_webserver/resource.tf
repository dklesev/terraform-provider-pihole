# Manage webserver settings
resource "pihole_config_webserver" "settings" {
  domain          = "pi.hole"
  port            = "80o,443os,[::]:80o,[::]:443os"
  threads         = 50
  serve_all       = false
  session_timeout = 1800  # 30 minutes
  session_restore = true

  # Interface settings
  interface_boxed = true
  interface_theme = "default-auto"
}
