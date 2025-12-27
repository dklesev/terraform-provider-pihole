// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccResourceGroupConfig("test-group-basic", true, "Basic test group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_group.test", "name", "test-group-basic"),
					resource.TestCheckResourceAttr("pihole_group.test", "enabled", "true"),
					resource.TestCheckResourceAttr("pihole_group.test", "description", "Basic test group"),
					resource.TestCheckResourceAttrSet("pihole_group.test", "id"),
					resource.TestCheckResourceAttrSet("pihole_group.test", "date_added"),
				),
			},
			// ImportState
			{
				ResourceName:      "pihole_group.test",
				ImportState:       true,
				ImportStateId:     "test-group-basic",
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccResourceGroupConfig("test-group-basic", false, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_group.test", "name", "test-group-basic"),
					resource.TestCheckResourceAttr("pihole_group.test", "enabled", "false"),
					resource.TestCheckResourceAttr("pihole_group.test", "description", "Updated description"),
				),
			},
			// Update name
			{
				Config: testAccResourceGroupConfig("test-group-renamed", false, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_group.test", "name", "test-group-renamed"),
				),
			},
		},
	})
}

func TestAccResourceGroup_minimal(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupConfigMinimal("test-group-minimal"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_group.test", "name", "test-group-minimal"),
					resource.TestCheckResourceAttr("pihole_group.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceGroup_disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupConfig("test-group-disabled", false, "Disabled group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_group.test", "enabled", "false"),
				),
			},
		},
	})
}

func testAccResourceGroupConfig(name string, enabled bool, description string) string {
	return fmt.Sprintf(`
resource "pihole_group" "test" {
  name        = %[1]q
  enabled     = %[2]t
  description = %[3]q
}
`, name, enabled, description)
}

func testAccResourceGroupConfigMinimal(name string) string {
	return fmt.Sprintf(`
resource "pihole_group" "test" {
  name = %[1]q
}
`, name)
}
