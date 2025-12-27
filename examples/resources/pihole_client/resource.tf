# Manage a client by IP address
resource "pihole_client" "desktop" {
  client  = "192.168.1.100"
  comment = "Office desktop"
}

# Manage a client by MAC address
resource "pihole_client" "laptop" {
  client  = "AA:BB:CC:DD:EE:FF"
  comment = "Work laptop"
}

# Manage an entire subnet
resource "pihole_client" "iot_subnet" {
  client  = "192.168.10.0/24"
  groups  = [pihole_group.iot.id]
  comment = "IoT network - strict blocking"
}
