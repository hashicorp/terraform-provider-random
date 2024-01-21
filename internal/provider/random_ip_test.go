// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceIP(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_ip" "basic" {
							address_type = "ipv4"
							cidr_range   = "10.0.0.0/16"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("random_ip.example", "address_type"),
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
