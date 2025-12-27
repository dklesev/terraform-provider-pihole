// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceClient_byIP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig("192.168.1.100", "Test client by IP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_client.test", "client", "192.168.1.100"),
					resource.TestCheckResourceAttr("pihole_client.test", "comment", "Test client by IP"),
					resource.TestCheckResourceAttrSet("pihole_client.test", "id"),
				),
			},
			{
				ResourceName:      "pihole_client.test",
				ImportState:       true,
				ImportStateId:     "192.168.1.100",
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccResourceClientConfig("192.168.1.100", "Updated comment"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_client.test", "comment", "Updated comment"),
				),
			},
		},
	})
}

func TestAccResourceClient_byMAC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig("AA:BB:CC:DD:EE:FF", "Test client by MAC"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_client.test", "client", "AA:BB:CC:DD:EE:FF"),
				),
			},
		},
	})
}

func TestAccResourceClient_bySubnet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig("192.168.10.0/24", "Test client by subnet"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_client.test", "client", "192.168.10.0/24"),
				),
			},
		},
	})
}

func TestAccResourceClient_withGroups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientWithGroupConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_client.test", "client", "192.168.1.200"),
					resource.TestCheckResourceAttr("pihole_client.test", "groups.#", "1"),
				),
			},
		},
	})
}

func testAccResourceClientConfig(client, comment string) string {
	return fmt.Sprintf(`
resource "pihole_client" "test" {
  client  = %[1]q
  comment = %[2]q
}
`, client, comment)
}

func testAccResourceClientWithGroupConfig() string {
	return `
resource "pihole_group" "test" {
  name = "client-test-group"
}

resource "pihole_client" "test" {
  client  = "192.168.1.200"
  groups  = [pihole_group.test.id]
  comment = "Client with group"
}
`
}
