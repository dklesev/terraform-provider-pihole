// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceConfigDNS_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccResourceConfigDNSBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_dns.test", "port", "53"),
					resource.TestCheckResourceAttr("pihole_config_dns.test", "dnssec", "false"),
					resource.TestCheckResourceAttr("pihole_config_dns.test", "query_logging", "true"),
					resource.TestCheckResourceAttr("pihole_config_dns.test", "blocking_active", "true"),
					resource.TestCheckResourceAttrSet("pihole_config_dns.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "pihole_config_dns.test",
				ImportState:       true,
				ImportStateId:     "dns",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceConfigDNS_dnssec(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Enable DNSSEC
			{
				Config: testAccResourceConfigDNSDNSSEC(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_dns.test", "dnssec", "true"),
				),
			},
			// Disable DNSSEC
			{
				Config: testAccResourceConfigDNSDNSSEC(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_dns.test", "dnssec", "false"),
				),
			},
		},
	})
}

func TestAccResourceConfigDNS_cacheSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Default cache settings
			{
				Config: testAccResourceConfigDNSCache(10000, 3600),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_dns.test", "cache_size", "10000"),
					resource.TestCheckResourceAttr("pihole_config_dns.test", "cache_optimizer", "3600"),
				),
			},
			// Larger cache
			{
				Config: testAccResourceConfigDNSCache(50000, 7200),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_dns.test", "cache_size", "50000"),
					resource.TestCheckResourceAttr("pihole_config_dns.test", "cache_optimizer", "7200"),
				),
			},
		},
	})
}

func TestAccResourceConfigDNS_rateLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfigDNSRateLimit(1000, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_dns.test", "rate_limit_count", "1000"),
					resource.TestCheckResourceAttr("pihole_config_dns.test", "rate_limit_interval", "60"),
				),
			},
			// Stricter rate limit
			{
				Config: testAccResourceConfigDNSRateLimit(500, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_dns.test", "rate_limit_count", "500"),
					resource.TestCheckResourceAttr("pihole_config_dns.test", "rate_limit_interval", "30"),
				),
			},
		},
	})
}

// Test config helpers

func testAccResourceConfigDNSBasic() string {
	return `
resource "pihole_config_dns" "test" {
  dnssec        = false
  query_logging = true
}
`
}

func testAccResourceConfigDNSDNSSEC(enabled bool) string {
	enabledStr := "false"
	if enabled {
		enabledStr = "true"
	}
	return `
resource "pihole_config_dns" "test" {
  dnssec = ` + enabledStr + `
}
`
}

func testAccResourceConfigDNSCache(size, optimizer int) string {
	return `
resource "pihole_config_dns" "test" {
  cache_size      = ` + itoa(size) + `
  cache_optimizer = ` + itoa(optimizer) + `
}
`
}

func testAccResourceConfigDNSRateLimit(count, interval int) string {
	return `
resource "pihole_config_dns" "test" {
  rate_limit_count    = ` + itoa(count) + `
  rate_limit_interval = ` + itoa(interval) + `
}
`
}

// Helper function
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}
