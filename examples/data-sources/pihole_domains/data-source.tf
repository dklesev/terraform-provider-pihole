# Retrieve all domains
data "pihole_domains" "all" {}

# Retrieve only deny domains
data "pihole_domains" "blocked" {
  type = "deny"
}

# Retrieve only exact match allow domains
data "pihole_domains" "exact_allow" {
  type = "allow"
  kind = "exact"
}
