// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDNSBlocking_enable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDNSBlockingConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_dns_blocking.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceDNSBlocking_disable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDNSBlockingConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_dns_blocking.test", "enabled", "false"),
				),
			},
			// Toggle back to enabled
			{
				Config: testAccResourceDNSBlockingConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_dns_blocking.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceDNSBlocking_withTimer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDNSBlockingWithTimerConfig(false, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_dns_blocking.test", "enabled", "false"),
					// Note: Timer value is racy since it counts down in real-time
					resource.TestCheckResourceAttrSet("pihole_dns_blocking.test", "timer"),
				),
			},
		},
	})
}

func testAccResourceDNSBlockingConfig(enabled bool) string {
	if enabled {
		return `
resource "pihole_dns_blocking" "test" {
  enabled = true
}
`
	}
	return `
resource "pihole_dns_blocking" "test" {
  enabled = false
}
`
}

func testAccResourceDNSBlockingWithTimerConfig(enabled bool, timer int) string {
	return `
resource "pihole_dns_blocking" "test" {
  enabled = false
  timer   = 300
}
`
}
