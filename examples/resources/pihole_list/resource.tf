# Subscribe to a popular blocklist
resource "pihole_list" "stevenblack" {
  address = "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts"
  type    = "block"
  enabled = true
  comment = "StevenBlack unified hosts"
}

# Add an allowlist
resource "pihole_list" "whitelist" {
  address = "https://raw.githubusercontent.com/anudeepND/whitelist/master/domains/whitelist.txt"
  type    = "allow"
  enabled = true
  comment = "Community maintained whitelist"
}

# Blocklist for specific group only
resource "pihole_list" "iot_blocklist" {
  address = "https://example.com/iot-blocklist.txt"
  type    = "block"
  enabled = true
  groups  = [pihole_group.iot.id]
  comment = "Extra blocklist for IoT devices"
}
