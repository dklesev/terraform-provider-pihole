// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDomain_exactDeny(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create exact deny
			{
				Config: testAccResourceDomainConfig("test.example.com", "deny", "exact", true, "Exact deny test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_domain.test", "domain", "test.example.com"),
					resource.TestCheckResourceAttr("pihole_domain.test", "type", "deny"),
					resource.TestCheckResourceAttr("pihole_domain.test", "kind", "exact"),
					resource.TestCheckResourceAttr("pihole_domain.test", "enabled", "true"),
					resource.TestCheckResourceAttr("pihole_domain.test", "comment", "Exact deny test"),
					resource.TestCheckResourceAttrSet("pihole_domain.test", "id"),
				),
			},
			// Import
			{
				ResourceName:      "pihole_domain.test",
				ImportState:       true,
				ImportStateId:     "deny/exact/test.example.com",
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccResourceDomainConfig("test.example.com", "deny", "exact", false, "Updated comment"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_domain.test", "enabled", "false"),
					resource.TestCheckResourceAttr("pihole_domain.test", "comment", "Updated comment"),
				),
			},
		},
	})
}

func TestAccResourceDomain_exactAllow(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDomainConfig("allowed.example.com", "allow", "exact", true, "Exact allow test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_domain.test", "domain", "allowed.example.com"),
					resource.TestCheckResourceAttr("pihole_domain.test", "type", "allow"),
					resource.TestCheckResourceAttr("pihole_domain.test", "kind", "exact"),
				),
			},
			{
				ResourceName:      "pihole_domain.test",
				ImportState:       true,
				ImportStateId:     "allow/exact/allowed.example.com",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDomain_regexDeny(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDomainConfig("^ads\\..*\\.example\\.com$", "deny", "regex", true, "Regex deny test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_domain.test", "domain", "^ads\\..*\\.example\\.com$"),
					resource.TestCheckResourceAttr("pihole_domain.test", "type", "deny"),
					resource.TestCheckResourceAttr("pihole_domain.test", "kind", "regex"),
				),
			},
		},
	})
}

func TestAccResourceDomain_regexAllow(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDomainConfig("^.*\\.trusted\\.com$", "allow", "regex", true, "Regex allow test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_domain.test", "domain", "^.*\\.trusted\\.com$"),
					resource.TestCheckResourceAttr("pihole_domain.test", "type", "allow"),
					resource.TestCheckResourceAttr("pihole_domain.test", "kind", "regex"),
				),
			},
		},
	})
}

func TestAccResourceDomain_withGroups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDomainWithGroupConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_domain.test", "domain", "grouped.example.com"),
					resource.TestCheckResourceAttr("pihole_domain.test", "groups.#", "1"),
				),
			},
		},
	})
}

func testAccResourceDomainConfig(domain, domainType, kind string, enabled bool, comment string) string {
	return fmt.Sprintf(`
resource "pihole_domain" "test" {
  domain  = %[1]q
  type    = %[2]q
  kind    = %[3]q
  enabled = %[4]t
  comment = %[5]q
}
`, domain, domainType, kind, enabled, comment)
}

func testAccResourceDomainWithGroupConfig() string {
	return `
resource "pihole_group" "test" {
  name = "domain-test-group"
}

resource "pihole_domain" "test" {
  domain  = "grouped.example.com"
  type    = "deny"
  kind    = "exact"
  enabled = true
  groups  = [pihole_group.test.id]
}
`
}
