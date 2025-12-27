// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceDomains_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDomainsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pihole_domains.test", "domains.#"),
				),
			},
		},
	})
}

func TestAccDataSourceDomains_filterByType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDomainsFilterByTypeConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pihole_domains.test", "domains.#"),
				),
			},
		},
	})
}

func TestAccDataSourceDomains_filterByKind(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDomainsFilterByKindConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pihole_domains.test", "domains.#"),
				),
			},
		},
	})
}

func testAccDataSourceDomainsConfig() string {
	return `
resource "pihole_domain" "test" {
  domain  = "ds-test.example.com"
  type    = "deny"
  kind    = "exact"
  enabled = true
}

data "pihole_domains" "test" {
  depends_on = [pihole_domain.test]
}
`
}

func testAccDataSourceDomainsFilterByTypeConfig() string {
	return `
resource "pihole_domain" "test" {
  domain  = "ds-filter-type.example.com"
  type    = "deny"
  kind    = "exact"
  enabled = true
}

data "pihole_domains" "test" {
  type       = "deny"
  depends_on = [pihole_domain.test]
}
`
}

func testAccDataSourceDomainsFilterByKindConfig() string {
	return `
resource "pihole_domain" "test" {
  domain  = "ds-filter-kind.example.com"
  type    = "deny"
  kind    = "exact"
  enabled = true
}

data "pihole_domains" "test" {
  kind       = "exact"
  depends_on = [pihole_domain.test]
}
`
}
