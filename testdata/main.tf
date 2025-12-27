# Test Pi-hole provider configuration
# Apply this against the docker Pi-hole to verify everything works

terraform {
  required_providers {
    pihole = {
      source  = "dklesev/pihole"
      version = "99.0.0"
    }
  }
}

provider "pihole" {
  url                    = "http://localhost:8080"
  tls_insecure_skip_verify = true
}

locals {
  config = yamldecode(file("${path.module}/config_full.yaml"))
}
