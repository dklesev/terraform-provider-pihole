// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceLists_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceListsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pihole_lists.test", "lists.#"),
				),
			},
		},
	})
}

func TestAccDataSourceLists_filterByType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceListsFilterConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pihole_lists.test", "lists.#"),
				),
			},
		},
	})
}

func testAccDataSourceListsConfig() string {
	return `
resource "pihole_list" "test" {
  address = "https://example.com/ds-test-list.txt"
  type    = "block"
  enabled = true
}

data "pihole_lists" "test" {
  depends_on = [pihole_list.test]
}
`
}

func testAccDataSourceListsFilterConfig() string {
	return `
resource "pihole_list" "test" {
  address = "https://example.com/ds-filter-list.txt"
  type    = "block"
  enabled = true
}

data "pihole_lists" "test" {
  type       = "block"
  depends_on = [pihole_list.test]
}
`
}
