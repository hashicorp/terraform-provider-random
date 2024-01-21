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
				Config: `resource "random_ip" "test" {
							address_type = "ipv4"
							cidr_range   = "10.0.0.0/16"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_ip.test", "address_type", "ipv4"),
					resource.TestCheckResourceAttrSet("random_ip.test", "cidr_range"),
					resource.TestCheckResourceAttrSet("random_ip.test", "result"),
				),
			},
		},
	})
}

func TestAccResourceIP_ipv6(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_ip" "test" {
							address_type = "ipv6"
							cidr_range   = "2001:db8::/16"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_ip.test", "address_type", "ipv6"),
					resource.TestCheckResourceAttrSet("random_ip.test", "cidr_range"),
					resource.TestCheckResourceAttrSet("random_ip.test", "result"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Keep_EmptyMap(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers      = {}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers      = {}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesEqual(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Keep_EmptyMapToNullValue(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers      = {}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = null
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesEqual(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Keep_NullMap(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesEqual(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Keep_NullMapToNullValue(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = null
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesEqual(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Keep_NullValue(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = null
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = null
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesEqual(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Keep_NullValues(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key1" = null
						"key2" = null
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "2"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key1" = null
						"key2" = null
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesEqual(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "2"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Keep_Value(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = "123"
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = "123"
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesEqual(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Keep_Values(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "2"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesEqual(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "2"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Replace_EmptyMapToValue(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers      = {}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = "123"
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesDiffer(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Replace_NullMapToValue(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = "123"
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesDiffer(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Replace_NullValueToValue(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = null
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = "123"
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesDiffer(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Replace_ValueToEmptyMap(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = "123"
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers      = {}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesDiffer(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Replace_ValueToNullMap(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = "123"
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesDiffer(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Replace_ValueToNullValue(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = "123"
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = null
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesDiffer(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourceIP_Keepers_Replace_ValueToNewValue(t *testing.T) {
	var id1, id2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = "123"
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					keepers = {
						"key" = "456"
					}
					address_type = "ipv4"
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					testCheckAttributeValuesDiffer(&id1, &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
		},
	})
}
