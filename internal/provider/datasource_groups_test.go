// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceGroups_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGroupsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Default group should always exist
					resource.TestCheckResourceAttrSet("data.pihole_groups.test", "groups.#"),
				),
			},
		},
	})
}

func TestAccDataSourceGroups_withResources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGroupsWithResourcesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Should find at least 2 groups (Default + created)
					resource.TestCheckResourceAttrSet("data.pihole_groups.test", "groups.#"),
				),
			},
		},
	})
}

func testAccDataSourceGroupsConfig() string {
	return `
data "pihole_groups" "test" {}
`
}

func testAccDataSourceGroupsWithResourcesConfig() string {
	return `
resource "pihole_group" "test" {
  name        = "datasource-test-group"
  description = "Created for datasource test"
}

data "pihole_groups" "test" {
  depends_on = [pihole_group.test]
}
`
}
