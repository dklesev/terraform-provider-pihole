// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceList_blocklist(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceListConfig("https://block.example.com/acc-test-blocklist.txt", "block", true, "ACC test blocklist"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_list.test", "address", "https://block.example.com/acc-test-blocklist.txt"),
					resource.TestCheckResourceAttr("pihole_list.test", "type", "block"),
					resource.TestCheckResourceAttr("pihole_list.test", "enabled", "true"),
					resource.TestCheckResourceAttr("pihole_list.test", "comment", "ACC test blocklist"),
					resource.TestCheckResourceAttrSet("pihole_list.test", "id"),
				),
			},
			{
				ResourceName:      "pihole_list.test",
				ImportState:       true,
				ImportStateId:     "block/https://block.example.com/acc-test-blocklist.txt",
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccResourceListConfig("https://block.example.com/acc-test-blocklist.txt", "block", false, "Disabled blocklist"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_list.test", "enabled", "false"),
					resource.TestCheckResourceAttr("pihole_list.test", "comment", "Disabled blocklist"),
				),
			},
		},
	})
}

func TestAccResourceList_allowlist(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceListConfig("https://raw.githubusercontent.com/anudeepND/whitelist/master/domains/whitelist.txt", "allow", true, "Community whitelist"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_list.test", "type", "allow"),
					resource.TestCheckResourceAttr("pihole_list.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceList_withGroups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceListWithGroupConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_list.test", "groups.#", "1"),
				),
			},
		},
	})
}

func testAccResourceListConfig(address, listType string, enabled bool, comment string) string {
	return fmt.Sprintf(`
resource "pihole_list" "test" {
  address = %[1]q
  type    = %[2]q
  enabled = %[3]t
  comment = %[4]q
}
`, address, listType, enabled, comment)
}

func testAccResourceListWithGroupConfig() string {
	return `
resource "pihole_group" "test" {
  name = "list-test-group"
}

resource "pihole_list" "test" {
  address = "https://example.com/test-list.txt"
  type    = "block"
  enabled = true
  groups  = [pihole_group.test.id]
  comment = "List with group"
}
`
}
