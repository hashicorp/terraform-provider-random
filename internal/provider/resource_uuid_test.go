// Copyright IBM Corp. 2017, 2026
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

func TestAccResourceUUID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_uuid" "basic" { 
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_uuid.basic", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}`))),
				},
			},
			{
				ResourceName:      "random_uuid.basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceUUID_ImportWithoutKeepersProducesNoPlannedChanges(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_uuid" "basic" { 
						}`,
				ResourceName:       "random_uuid.basic",
				ImportStateId:      "6b0f8e7c-3ea6-4523-88a2-5a70419ee954",
				ImportState:        true,
				ImportStatePersist: true,
			},
			{
				Config: `resource "random_uuid" "basic" { 
						}`,
				PlanOnly: true,
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Keep_EmptyMap(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Keep_EmptyMapToNullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Keep_NullMap(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Keep_NullMapToNullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Keep_NullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Keep_NullValues(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Keep_Value(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Keep_Values(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Replace_EmptyMapToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Replace_NullMapToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Replace_NullValueToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Replace_ValueToEmptyMap(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Replace_ValueToNullMap(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Replace_ValueToNullValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_Replace_ValueToNewValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_FrameworkMigration_NullMapToNullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_FrameworkMigration_NullMapToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_FrameworkMigration_NullMapToMultipleNullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_FrameworkMigration_NullMapToMultipleValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_FrameworkMigration_NullMapValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceUUID_Keepers_FrameworkMigration_NullMapValueToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "test" {
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_uuid.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_uuid.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceUUID_UpgradeFromVersion3_3_2(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_uuid" "basic" { 
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_uuid.basic", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}`))),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "basic" { 
						}`,
				PlanOnly: true,
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_uuid" "basic" { 
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_uuid.basic", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}`))),
				},
			},
		},
	})
}
