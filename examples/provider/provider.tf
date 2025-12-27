terraform {
  required_providers {
    pihole = {
      source = "registry.terraform.io/dklesev/pihole"
    }
  }
}

# Configure the Pi-hole provider
provider "pihole" {
  # URL of your Pi-hole instance
  # Can also be set via PIHOLE_URL environment variable
  url = "http://pi.hole"

  # Password for the Pi-hole web interface
  # Can also be set via PIHOLE_PASSWORD environment variable
  password = "your-password"

  # Optional: Skip TLS verification (not recommended for production)
  # tls_insecure_skip_verify = true

  # Optional: HTTP timeout in seconds (default: 30)
  # timeout = 60
}
