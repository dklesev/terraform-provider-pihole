# Retrieve all Pi-hole groups
data "pihole_groups" "all" {}

# Output group names
output "group_names" {
  value = [for g in data.pihole_groups.all.groups : g.name]
}
