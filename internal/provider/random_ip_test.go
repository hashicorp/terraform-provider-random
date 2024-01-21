// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceIP_ipv4(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_ip" "basic" {
							address_type = "ipv4"
							cidr_range   = "10.0.0.0/16"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_ip.example", "address_type", "ipv4"),
					resource.TestCheckResourceAttrSet("random_ip.example", "cidr_range"),
					resource.TestCheckResourceAttrSet("random_ip.example", "result"),
				),
			},
			{
				ResourceName:      "random_ip.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceIP_ipv6(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_ip" "basic" {
							address_type = "ipv6"
							cidr_range   = "2001:db8::/16"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_ip.example", "address_type", "ipv6"),
					resource.TestCheckResourceAttrSet("random_ip.example", "cidr_range"),
					resource.TestCheckResourceAttrSet("random_ip.example", "result"),
				),
			},
			{
				ResourceName:      "random_ip.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
