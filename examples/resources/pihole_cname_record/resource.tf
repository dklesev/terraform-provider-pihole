# Manage CNAME records
resource "pihole_cname_record" "www" {
  domain = "www.example.local"
  target = "server.example.local"
}

resource "pihole_cname_record" "api" {
  domain = "api.example.local"
  target = "server.example.local"
}
