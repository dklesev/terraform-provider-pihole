// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceConfigMisc_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccResourceConfigMiscBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "privacy_level", "0"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "etc_dnsmasq_d", "false"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "extra_logging", "false"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "read_only", "false"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "check_load", "true"),
					resource.TestCheckResourceAttrSet("pihole_config_misc.test", "nice"),
				),
			},
			// ImportState
			{
				ResourceName:      "pihole_config_misc.test",
				ImportState:       true,
				ImportStateId:     "misc",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceConfigMisc_withDnsmasqLines(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with dnsmasq_lines
			{
				Config: testAccResourceConfigMiscWithDnsmasqLines([]string{
					"address=/test1.local/192.168.1.100",
					"address=/test2.local/192.168.1.101",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "dnsmasq_lines.#", "2"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "dnsmasq_lines.0", "address=/test1.local/192.168.1.100"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "dnsmasq_lines.1", "address=/test2.local/192.168.1.101"),
				),
			},
			// Update dnsmasq_lines - add one more
			{
				Config: testAccResourceConfigMiscWithDnsmasqLines([]string{
					"address=/test1.local/192.168.1.100",
					"address=/test2.local/192.168.1.101",
					"address=/test3.local/192.168.1.102",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "dnsmasq_lines.#", "3"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "dnsmasq_lines.2", "address=/test3.local/192.168.1.102"),
				),
			},
			// Update dnsmasq_lines - remove one
			{
				Config: testAccResourceConfigMiscWithDnsmasqLines([]string{
					"address=/test1.local/192.168.1.100",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "dnsmasq_lines.#", "1"),
				),
			},
			// Clear dnsmasq_lines
			{
				Config: testAccResourceConfigMiscWithDnsmasqLines([]string{}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "dnsmasq_lines.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceConfigMisc_privacyLevels(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Privacy level 0 (show everything)
			{
				Config: testAccResourceConfigMiscPrivacyLevel(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "privacy_level", "0"),
				),
			},
			// Privacy level 1
			{
				Config: testAccResourceConfigMiscPrivacyLevel(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "privacy_level", "1"),
				),
			},
			// Privacy level 2
			{
				Config: testAccResourceConfigMiscPrivacyLevel(2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "privacy_level", "2"),
				),
			},
			// Privacy level 3 (hide everything)
			{
				Config: testAccResourceConfigMiscPrivacyLevel(3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "privacy_level", "3"),
				),
			},
			// Back to 0
			{
				Config: testAccResourceConfigMiscPrivacyLevel(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "privacy_level", "0"),
				),
			},
		},
	})
}

func TestAccResourceConfigMisc_checkSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Default check settings
			{
				Config: testAccResourceConfigMiscCheckSettings(true, 90, 90),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "check_load", "true"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "check_shmem", "90"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "check_disk", "90"),
				),
			},
			// Disable load checking
			{
				Config: testAccResourceConfigMiscCheckSettings(false, 90, 90),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "check_load", "false"),
				),
			},
			// Lower thresholds
			{
				Config: testAccResourceConfigMiscCheckSettings(true, 80, 75),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "check_load", "true"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "check_shmem", "80"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "check_disk", "75"),
				),
			},
		},
	})
}

func TestAccResourceConfigMisc_allFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfigMiscAllFields(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pihole_config_misc.test", "privacy_level", "1"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "delay_startup", "5"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "nice", "-5"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "addr2line", "true"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "etc_dnsmasq_d", "false"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "extra_logging", "true"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "read_only", "false"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "normalize_cpu", "true"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "hide_dnsmasq_warn", "false"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "check_load", "true"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "check_shmem", "85"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "check_disk", "80"),
					resource.TestCheckResourceAttr("pihole_config_misc.test", "dnsmasq_lines.#", "2"),
				),
			},
		},
	})
}

// Test config helpers

func testAccResourceConfigMiscBasic() string {
	return `
resource "pihole_config_misc" "test" {
  privacy_level = 0
  etc_dnsmasq_d = false
}
`
}

func testAccResourceConfigMiscWithDnsmasqLines(lines []string) string {
	if len(lines) == 0 {
		return `
resource "pihole_config_misc" "test" {
  dnsmasq_lines = []
}
`
	}

	linesStr := ""
	for _, line := range lines {
		linesStr += `    "` + line + `",` + "\n"
	}

	return `
resource "pihole_config_misc" "test" {
  dnsmasq_lines = [
` + linesStr + `  ]
}
`
}

func testAccResourceConfigMiscPrivacyLevel(level int) string {
	return fmt.Sprintf(`
resource "pihole_config_misc" "test" {
  privacy_level = %d
}
`, level)
}

func testAccResourceConfigMiscCheckSettings(checkLoad bool, checkShmem, checkDisk int) string {
	return fmt.Sprintf(`
resource "pihole_config_misc" "test" {
  check_load  = %t
  check_shmem = %d
  check_disk  = %d
}
`, checkLoad, checkShmem, checkDisk)
}

func testAccResourceConfigMiscAllFields() string {
	return `
resource "pihole_config_misc" "test" {
  privacy_level     = 1
  delay_startup     = 5
  nice              = -5
  addr2line         = true
  etc_dnsmasq_d     = false
  extra_logging     = true
  read_only         = false
  normalize_cpu     = true
  hide_dnsmasq_warn = false
  check_load        = true
  check_shmem       = 85
  check_disk        = 80
  dnsmasq_lines     = [
    "address=/test.local/192.168.1.100",
    "server=/corp.local/10.0.0.1"
  ]
}
`
}
