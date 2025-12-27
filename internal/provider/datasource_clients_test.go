// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceClients_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceClientsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pihole_clients.test", "clients.#"),
				),
			},
		},
	})
}

func testAccDataSourceClientsConfig() string {
	return `
resource "pihole_client" "test" {
  client  = "192.168.1.250"
  comment = "Datasource test client"
}

data "pihole_clients" "test" {
  depends_on = [pihole_client.test]
}
`
}
