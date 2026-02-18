// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// These results are current as of Go 1.6. The Go
// "rand" package does not guarantee that the random
// number generator will generate the same results
// forever, but the maintainers endeavor not to change
// it gratuitously.
// These tests allow us to detect such changes and
// document them when they arise, but the docs for this
// resource specifically warn that results are not
// guaranteed consistent across Terraform releases.
func TestAccResourceShuffle(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_shuffle" "default_length" {
    						input = ["a", "b", "c", "d", "e"]
    						seed = "-"
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_shuffle.default_length", tfjsonpath.New("result"),
						knownvalue.ListExact(
							[]knownvalue.Check{
								knownvalue.StringExact("a"),
								knownvalue.StringExact("c"),
								knownvalue.StringExact("b"),
								knownvalue.StringExact("e"),
								knownvalue.StringExact("d"),
							},
						),
					),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Keep_EmptyMap(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Keep_EmptyMapToNullValue(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Keep_NullMap(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Keep_NullMapToNullValue(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Keep_NullValue(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Keep_NullValues(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Keep_Value(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Keep_Values(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Replace_EmptyMapToValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Replace_NullMapToValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Replace_NullValueToValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Replace_ValueToEmptyMap(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Replace_ValueToNullMap(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Replace_ValueToNullValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_Replace_ValueToNewValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_FrameworkMigration_NullMapToNullValue(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_FrameworkMigration_NullMapToValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_FrameworkMigration_NullMapToMultipleNullValue(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_FrameworkMigration_NullMapToMultipleValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_FrameworkMigration_NullMapValue(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Keepers_FrameworkMigration_NullMapValueToValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "test" {
					input = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_shuffle.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-random/issues/409
func TestAccResourceShuffle_ResultCount_Zero(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_shuffle" "test" {
    						input        = ["a", "b", "c", "d", "e"]
    						result_count = 0
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_shuffle.test", tfjsonpath.New("result"), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_ResultCount_Shorter(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_shuffle" "shorter_length" {
    						input = ["a", "b", "c", "d", "e"]
    						seed = "-"
    						result_count = 3
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_shuffle.shorter_length", tfjsonpath.New("result"),
						knownvalue.ListExact(
							[]knownvalue.Check{
								knownvalue.StringExact("a"),
								knownvalue.StringExact("c"),
								knownvalue.StringExact("b"),
							},
						),
					),
				},
			},
		},
	})
}

func TestAccResourceShuffle_ResultCount_Longer(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_shuffle" "longer_length" {
    						input = ["a", "b", "c", "d", "e"]
    						seed = "-"
    						result_count = 12
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_shuffle.longer_length", tfjsonpath.New("result"),
						knownvalue.ListExact(
							[]knownvalue.Check{
								knownvalue.StringExact("a"),
								knownvalue.StringExact("c"),
								knownvalue.StringExact("b"),
								knownvalue.StringExact("e"),
								knownvalue.StringExact("d"),
								knownvalue.StringExact("a"),
								knownvalue.StringExact("e"),
								knownvalue.StringExact("d"),
								knownvalue.StringExact("c"),
								knownvalue.StringExact("b"),
								knownvalue.StringExact("a"),
								knownvalue.StringExact("b"),
							},
						),
					),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Input_Empty(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_shuffle" "empty_length" {
    						input = []
    						seed = "-"
    						result_count = 12
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_shuffle.empty_length", tfjsonpath.New("result"), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourceShuffle_Input_One(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_shuffle" "one_length" {
    						input = ["a"]
    						seed = "-"
    						result_count = 1
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_shuffle.one_length", tfjsonpath.New("result"),
						knownvalue.ListExact(
							[]knownvalue.Check{
								knownvalue.StringExact("a"),
							},
						),
					),
				},
			},
		},
	})
}

func TestAccResourceShuffle_UpgradeFromVersion3_3_2(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_shuffle" "default_length" {
    						input = ["a", "b", "c", "d", "e"]
    						seed = "-"
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_shuffle.default_length", tfjsonpath.New("result"),
						knownvalue.ListExact(
							[]knownvalue.Check{
								knownvalue.StringExact("a"),
								knownvalue.StringExact("c"),
								knownvalue.StringExact("b"),
								knownvalue.StringExact("e"),
								knownvalue.StringExact("d"),
							},
						),
					),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "default_length" {
    						input = ["a", "b", "c", "d", "e"]
    						seed = "-"
						}`,
				PlanOnly: true,
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "default_length" {
    						input = ["a", "b", "c", "d", "e"]
    						seed = "-"
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_shuffle.default_length", tfjsonpath.New("result"),
						knownvalue.ListExact(
							[]knownvalue.Check{
								knownvalue.StringExact("a"),
								knownvalue.StringExact("c"),
								knownvalue.StringExact("b"),
								knownvalue.StringExact("e"),
								knownvalue.StringExact("d"),
							},
						),
					),
				},
			},
		},
	})
}
