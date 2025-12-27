# Retrieve all configured clients
data "pihole_clients" "all" {}

# Output client identifiers
output "client_ids" {
  value = [for c in data.pihole_clients.all.clients : c.client]
}
