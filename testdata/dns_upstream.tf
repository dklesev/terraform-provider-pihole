# DNS Upstream servers

resource "pihole_dns_upstream" "cloudflare" {
  upstream = "1.1.1.1"
}

resource "pihole_dns_upstream" "quad9" {
  upstream = "9.9.9.9"
}
