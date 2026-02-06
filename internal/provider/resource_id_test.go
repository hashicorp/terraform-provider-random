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
	"github.com/terraform-providers/terraform-provider-random/internal/randomtest"
)

func TestAccResourceID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_id" "foo" {
  							byte_length = 4
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_id.foo", tfjsonpath.New("b64_url"), randomtest.StringLengthExact(6)),
					statecheck.ExpectKnownValue("random_id.foo", tfjsonpath.New("b64_std"), randomtest.StringLengthExact(8)),
					statecheck.ExpectKnownValue("random_id.foo", tfjsonpath.New("hex"), randomtest.StringLengthExact(8)),
					statecheck.ExpectKnownValue("random_id.foo", tfjsonpath.New("dec"), randomtest.StringLengthMin(1)),
				},
			},
			{
				ResourceName:      "random_id.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceID_ImportWithPrefix(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_id" "bar" {
  							byte_length = 4
  							prefix      = "cloud-"
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_id.bar", tfjsonpath.New("b64_url"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_id.bar", tfjsonpath.New("b64_std"), randomtest.StringLengthExact(14)),
					statecheck.ExpectKnownValue("random_id.bar", tfjsonpath.New("hex"), randomtest.StringLengthExact(14)),
					statecheck.ExpectKnownValue("random_id.bar", tfjsonpath.New("dec"), randomtest.StringLengthMin(1)),
				},
			},
			{
				ResourceName:        "random_id.bar",
				ImportState:         true,
				ImportStateIdPrefix: "cloud-,",
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccResourceID_ImportWithoutKeepersProducesNoPlannedChanges(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_id" "foo" {
  							byte_length = 4
						}`,
				ResourceName:       "random_id.foo",
				ImportStateId:      "p-9hUg",
				ImportState:        true,
				ImportStatePersist: true,
			},
			{
				Config: `resource "random_id" "foo" {
  							byte_length = 4
						}`,
				PlanOnly: true,
			},
		},
	})
}

func TestAccResourceID_UpgradeFromVersion3_3_2(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_id" "bar" {
  							byte_length = 4
  							prefix      = "cloud-"
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_id.bar", tfjsonpath.New("b64_url"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_id.bar", tfjsonpath.New("b64_std"), randomtest.StringLengthExact(14)),
					statecheck.ExpectKnownValue("random_id.bar", tfjsonpath.New("hex"), randomtest.StringLengthExact(14)),
					statecheck.ExpectKnownValue("random_id.bar", tfjsonpath.New("dec"), randomtest.StringLengthMin(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "bar" {
  							byte_length = 4
  							prefix      = "cloud-"
						}`,
				PlanOnly: true,
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "bar" {
  							byte_length = 4
  							prefix      = "cloud-"
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_id.bar", tfjsonpath.New("b64_url"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_id.bar", tfjsonpath.New("b64_std"), randomtest.StringLengthExact(14)),
					statecheck.ExpectKnownValue("random_id.bar", tfjsonpath.New("hex"), randomtest.StringLengthExact(14)),
					statecheck.ExpectKnownValue("random_id.bar", tfjsonpath.New("dec"), randomtest.StringLengthMin(1)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Keep_EmptyMap(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Keep_EmptyMapToNullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Keep_NullMap(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Keep_NullMapToNullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Keep_NullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Keep_NullValues(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Keep_Value(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Keep_Values(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Replace_EmptyMapToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Replace_NullMapToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Replace_NullValueToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Replace_ValueToEmptyMap(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Replace_ValueToNullMap(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Replace_ValueToNullValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_Replace_ValueToNewValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_FrameworkMigration_NullMapToNullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_FrameworkMigration_NullMapToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_FrameworkMigration_NullMapToMultipleNullValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_FrameworkMigration_NullMapToMultipleValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_FrameworkMigration_NullMapValue(t *testing.T) {
	// The id attribute values should be the same between test steps
	assertIdSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdSame.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourceID_Keepers_FrameworkMigration_NullMapValueToValue(t *testing.T) {
	// The id attribute values should differ between test steps
	assertIdDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "test" {
					byte_length = 4
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertIdDiffer.AddStateValue("random_id.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("random_id.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}
