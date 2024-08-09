// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceBytes(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_bytes" "basic" {
							length = 32
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_bytes.basic", tfjsonpath.New("base64"), knownvalue.StringRegexp(regexp.MustCompile(`^[A-Za-z/+\d]{43}=$`))),
					statecheck.ExpectKnownValue("random_bytes.basic", tfjsonpath.New("hex"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-f\d]{64}$`))),
					statecheck.ExpectKnownValue("random_bytes.basic", tfjsonpath.New("length"), knownvalue.Int64Exact(32)),
				},
			},
			{
				// Usage of ImportStateIdFunc is required as the value passed to the `terraform import` command needs
				// to be the bytes encoded with base64.
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					resource := "random_bytes.basic"
					rs, ok := s.RootModule().Resources[resource]
					if !ok {
						return "", fmt.Errorf("not found: %s", resource)
					}
					if rs.Primary.Attributes["base64"] == "" {
						return "", fmt.Errorf("no base64 attribute is set")
					}

					return rs.Primary.Attributes["base64"], nil
				},
				ResourceName:                         "random_bytes.basic",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "base64",
			},
		},
	})
}

func TestAccResourceBytes_ImportWithoutKeepersThenUpdateShouldNotTriggerChange(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				ImportState:        true,
				ImportStateId:      "hkvbcU5f8qGysTFhkI4gzf3yRWC1jXW3aRLCNQFOtNw=",
				ImportStatePersist: true,
				ResourceName:       "random_bytes.basic",
				Config: `resource "random_bytes" "basic" {
							length = 32
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_bytes.basic", tfjsonpath.New("base64"), knownvalue.StringExact("hkvbcU5f8qGysTFhkI4gzf3yRWC1jXW3aRLCNQFOtNw=")),
					statecheck.ExpectKnownValue("random_bytes.basic", tfjsonpath.New("hex"), knownvalue.StringExact("864bdb714e5ff2a1b2b13161908e20cdfdf24560b58d75b76912c235014eb4dc")),
					statecheck.ExpectKnownValue("random_bytes.basic", tfjsonpath.New("length"), knownvalue.Int64Exact(32)),
				},
			},
			{
				Config: `resource "random_bytes" "basic" {
							length = 32
						}`,
				PlanOnly: true,
			},
		},
	})
}

func TestAccResourceBytes_LengthErrors(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_bytes" "invalid_length" {
							length = 0
						}`,
				ExpectError: regexp.MustCompile(`.*Attribute length value must be at least 1, got: 0`),
			},
		},
	})
}

func TestAccResourceBytes_Length_ForceReplacement(t *testing.T) {
	// The base64 attribute values should differ between test steps
	assertBase64Differ := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 1
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("length"), knownvalue.Int64Exact(1)),
					assertBase64Differ.AddStateValue("random_bytes.test", tfjsonpath.New("base64")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("hex"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-f\d]{2}$`))),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 2
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("length"), knownvalue.Int64Exact(2)),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("hex"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-f\d]{4}$`))),
					assertBase64Differ.AddStateValue("random_bytes.test", tfjsonpath.New("base64")),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Keep_EmptyMap(t *testing.T) {
	// The hex attribute values should be the same between test steps
	assertHexSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexSame.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexSame.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Keep_NullMap(t *testing.T) {
	// The hex attribute values should be the same between test steps
	assertHexSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexSame.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexSame.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Keep_NullValue(t *testing.T) {
	// The hex attribute values should be the same between test steps
	assertHexSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexSame.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexSame.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Keep_NullValues(t *testing.T) {
	// The hex attribute values should be the same between test steps
	assertHexSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexSame.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexSame.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Keep_Value(t *testing.T) {
	// The hex attribute values should be the same between test steps
	assertHexSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexSame.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexSame.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Keep_Values(t *testing.T) {
	// The hex attribute values should be the same between test steps
	assertHexSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexSame.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexSame.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Replace_EmptyMapToValue(t *testing.T) {
	// The hex attribute values should differ between test steps
	assertHexDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Replace_NullMapToValue(t *testing.T) {
	// The hex attribute values should differ between test steps
	assertHexDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Replace_NullValueToValue(t *testing.T) {
	// The hex attribute values should differ between test steps
	assertHexDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Replace_ValueToEmptyMap(t *testing.T) {
	// The hex attribute values should differ between test steps
	assertHexDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Replace_ValueToNullMap(t *testing.T) {
	// The hex attribute values should differ between test steps
	assertHexDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Replace_ValueToNullValue(t *testing.T) {
	// The hex attribute values should differ between test steps
	assertHexDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceBytes_Keepers_Replace_ValueToNewValue(t *testing.T) {
	// The hex attribute values should differ between test steps
	assertHexDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 12
					keepers = {
						"key" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertHexDiffer.AddStateValue("random_bytes.test", tfjsonpath.New("hex")),
					statecheck.ExpectKnownValue("random_bytes.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}
