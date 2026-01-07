// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccResourceUUIDV7(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_uuid7" "basic" { 
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_uuid7.basic", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}`))),
				},
			},
			{
				ResourceName:      "random_uuid7.basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceUUIDV7_ImportWithoutKeepersProducesNoPlannedChanges(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_uuid7" "basic" { 
						}`,
				ResourceName:       "random_uuid7.basic",
				ImportStateId:      "6b0f8e7c-3ea6-7523-88a2-5a70719ee957",
				ImportState:        true,
				ImportStatePersist: true,
			},
			{
				Config: `resource "random_uuid7" "basic" { 
						}`,
				PlanOnly: true,
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Keep_EmptyMap(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Keep_EmptyMapToNullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Keep_NullMap(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Keep_NullMapToNullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Keep_NullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Keep_NullValues(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Keep_Value(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Keep_Values(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key1" = "123"
						"key2" = "756"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key1" = "123"
						"key2" = "756"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Replace_EmptyMapToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Replace_NullMapToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Replace_NullValueToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Replace_ValueToEmptyMap(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Replace_ValueToNullMap(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Replace_ValueToNullValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUIDV7_Keepers_Replace_ValueToNewValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid7" "test" {
					keepers = {
						"key" = "756"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid7.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid7.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}
