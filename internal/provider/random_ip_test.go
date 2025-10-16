// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceIP_basic(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_ip" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("random_ip.test", "result"),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourceIP_ipv4(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_ip" "test" {
							cidr_range   = "10.0.0.0/16"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("random_ip.test", "cidr_range"),
					resource.TestCheckResourceAttrSet("random_ip.test", "result"),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourceIP_ipv4_quadZero(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_ip" "test" {
							cidr_range   = "0.0.0.0/0"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("random_ip.test", "cidr_range"),
					resource.TestCheckResourceAttrSet("random_ip.test", "result"),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourceIP_ipv4_largestPrefix(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_ip" "test" {
							cidr_range   = "0.0.0.0/0"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("random_ip.test", "cidr_range"),
					resource.TestCheckResourceAttrSet("random_ip.test", "result"),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourceIP_ipv4_smallestPrefix(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_ip" "test" {
							cidr_range   = "192.168.1.1/32"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("random_ip.test", "cidr_range"),
					resource.TestCheckResourceAttrSet("random_ip.test", "result"),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
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
							cidr_range   = "2001:db8::/16"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("random_ip.test", "cidr_range"),
					resource.TestCheckResourceAttrSet("random_ip.test", "result"),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourceIP_ipv6_zeroCompression(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_ip" "test" {
							cidr_range   = "::/0"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("random_ip.test", "cidr_range"),
					resource.TestCheckResourceAttrSet("random_ip.test", "result"),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourceIP_ipv6_largestPrefix(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_ip" "test" {
							cidr_range   = "::/0"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("random_ip.test", "cidr_range"),
					resource.TestCheckResourceAttrSet("random_ip.test", "result"),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourceIP_ipv6_smallestPrefix(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_ip" "test" {
							cidr_range   = "2001:db8::1/128"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("random_ip.test", "cidr_range"),
					resource.TestCheckResourceAttrSet("random_ip.test", "result"),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
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
					cidr_range   = "2001:db8::/32"
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
					cidr_range   = "2001:db8::/32"
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
					cidr_range   = "2001:db8::/32"
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
					cidr_range   = "2001:db8::/32"
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
					cidr_range   = "2001:db8::/32"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id1),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_ip" "test" {
					cidr_range   = "2001:db8::/32"
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
					cidr_range   = "2001:db8::/32"
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
					cidr_range   = "2001:db8::/32"
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
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
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
					cidr_range   = "2001:db8::/32"
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
					cidr_range   = "2001:db8::/32"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
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
					cidr_range   = "2001:db8::/32"
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
					cidr_range   = "2001:db8::/32"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
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
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
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
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
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
					cidr_range   = "2001:db8::/32"
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
					cidr_range   = "2001:db8::/32"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
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
					cidr_range   = "10.1.0.0/24"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_ip.test", "id", &id2),
					resource.TestCheckResourceAttr("random_ip.test", "keepers.%", "1"),
				),
			},
		},
	})
}
