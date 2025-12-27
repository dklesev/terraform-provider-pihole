// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"pihole": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// Check for required environment variables
	if v := os.Getenv("PIHOLE_URL"); v == "" {
		// Use default test URL if not set
		os.Setenv("PIHOLE_URL", "http://localhost:8080")
	}
	if v := os.Getenv("PIHOLE_PASSWORD"); v == "" {
		// Use default test password if not set
		os.Setenv("PIHOLE_PASSWORD", "test123")
	}
}
