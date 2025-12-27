# Create a new Pi-hole group
resource "pihole_group" "example" {
  name        = "my-custom-group"
  enabled     = true
  description = "A group for specific devices"
}

# Output the group ID
output "group_id" {
  value = pihole_group.example.id
}
