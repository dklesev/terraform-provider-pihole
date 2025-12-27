// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDNSUpstream_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "pihole_dns_upstream" "test" {
  upstream = "208.67.222.222"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_dns_upstream.test", "upstream", "208.67.222.222"),
					resource.TestCheckResourceAttr("pihole_dns_upstream.test", "id", "208.67.222.222"),
				),
			},
			{
				ResourceName:      "pihole_dns_upstream.test",
				ImportState:       true,
				ImportStateId:     "208.67.222.222",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceLocalDNS_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "pihole_local_dns" "test" {
  hostname = "test.local"
  ip       = "192.168.1.100"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_local_dns.test", "hostname", "test.local"),
					resource.TestCheckResourceAttr("pihole_local_dns.test", "ip", "192.168.1.100"),
				),
			},
		},
	})
}

func TestAccResourceCNAMERecord_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "pihole_cname_record" "test" {
  domain = "www.test.local"
  target = "server.test.local"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_cname_record.test", "domain", "www.test.local"),
					resource.TestCheckResourceAttr("pihole_cname_record.test", "target", "server.test.local"),
				),
			},
		},
	})
}

func TestAccResourceDHCPStaticLease_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "pihole_dhcp_static_lease" "test" {
  mac      = "AA:BB:CC:DD:EE:FF"
  ip       = "192.168.1.200"
  hostname = "testhost"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_dhcp_static_lease.test", "mac", "AA:BB:CC:DD:EE:FF"),
					resource.TestCheckResourceAttr("pihole_dhcp_static_lease.test", "ip", "192.168.1.200"),
					resource.TestCheckResourceAttr("pihole_dhcp_static_lease.test", "hostname", "testhost"),
				),
			},
		},
	})
}
