# Retrieve all lists
data "pihole_lists" "all" {}

# Retrieve only blocklists
data "pihole_lists" "blocklists" {
  type = "block"
}

# Output blocklist addresses
output "blocklists" {
  value = [for l in data.pihole_lists.blocklists.lists : l.address]
}
