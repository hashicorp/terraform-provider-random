// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	res "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"golang.org/x/crypto/bcrypt"

	"github.com/terraform-providers/terraform-provider-random/internal/random"
	"github.com/terraform-providers/terraform-provider-random/internal/randomtest"
)

func TestGenerateHash(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input random.StringParams
	}{
		"defaults": {
			input: random.StringParams{
				Length:  73, // Required
				Lower:   true,
				Numeric: true,
				Special: true,
				Upper:   true,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			randomBytes, err := random.CreateString(testCase.input)

			if err != nil {
				t.Fatalf("unexpected random.CreateString error: %s", err)
			}

			hash, err := generateHash(string(randomBytes))

			if err != nil {
				t.Fatalf("unexpected generateHash error: %s", err)
			}

			err = bcrypt.CompareHashAndPassword([]byte(hash), randomBytes)

			if err != nil {
				t.Fatalf("unexpected bcrypt.CompareHashAndPassword error: %s", err)
			}
		})
	}
}

func TestCreateString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input         random.StringParams
		expectedError error
	}{
		"input-false": {
			input: random.StringParams{
				Length:  16, // Required
				Lower:   false,
				Numeric: false,
				Special: false,
				Upper:   false,
			},
			expectedError: errors.New("the character set specified is empty"),
		},
	}

	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := random.CreateString(testCase.input)

			if diff := cmp.Diff(testCase.expectedError, err, equateErrorMessage); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAccResourcePassword_Import(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "basic" {
							length = 12
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.basic", tfjsonpath.New("result"), randomtest.StringLengthExact(12)),
				},
			},
			{
				ResourceName: "random_password.basic",
				// Usage of ImportStateIdFunc is required as the value passed to the `terraform import` command needs
				// to be the password itself, as the password resource sets ID to "none" and "result" to the password
				// supplied during import.
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					id := "random_password.basic"
					rs, ok := s.RootModule().Resources[id]
					if !ok {
						return "", fmt.Errorf("not found: %s", id)
					}
					if rs.Primary.ID == "" {
						return "", fmt.Errorf("no ID is set")
					}

					return rs.Primary.Attributes["result"], nil
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bcrypt_hash"},
			},
		},
	})
}

func TestAccResourcePassword_BcryptHash(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "test" {
							length = 73
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						"random_password.test", tfjsonpath.New("bcrypt_hash"),
						"random_password.test", tfjsonpath.New("result"),
						randomtest.BcryptHashMatch(),
					),
				},
			},
		},
	})
}

// TestAccResourcePassword_BcryptHash_FromVersion3_3_2 verifies behaviour when
// upgrading state from schema V2 to V3 without a bcrypt_hash update.
func TestAccResourcePassword_BcryptHash_FromVersion3_3_2(t *testing.T) {
	// The bcrypt_hash attribute values should be the same between test steps
	assertBcryptHashSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertBcryptHashSame.AddStateValue("random_password.test", tfjsonpath.New("bcrypt_hash")),
					statecheck.CompareValuePairs(
						"random_password.test", tfjsonpath.New("bcrypt_hash"),
						"random_password.test", tfjsonpath.New("result"),
						randomtest.BcryptHashMatch(),
					),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertBcryptHashSame.AddStateValue("random_password.test", tfjsonpath.New("bcrypt_hash")),
					statecheck.CompareValuePairs(
						"random_password.test", tfjsonpath.New("bcrypt_hash"),
						"random_password.test", tfjsonpath.New("result"),
						randomtest.BcryptHashMatch(),
					),
				},
			},
		},
	})
}

// TestAccResourcePassword_BcryptHash_FromVersion3_4_2 verifies behaviour when
// upgrading state from schema V2 to V3 with an expected bcrypt_hash update.
func TestAccResourcePassword_BcryptHash_FromVersion3_4_2(t *testing.T) {
	// The bcrypt_hash attribute values should differ between test steps
	assertBcryptHashDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion342(),
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertBcryptHashDiffer.AddStateValue("random_password.test", tfjsonpath.New("bcrypt_hash")),
					statecheck.CompareValuePairs(
						"random_password.test", tfjsonpath.New("bcrypt_hash"),
						"random_password.test", tfjsonpath.New("result"),
						randomtest.BcryptHashMismatch(),
					),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertBcryptHashDiffer.AddStateValue("random_password.test", tfjsonpath.New("bcrypt_hash")),
					statecheck.CompareValuePairs(
						"random_password.test", tfjsonpath.New("bcrypt_hash"),
						"random_password.test", tfjsonpath.New("result"),
						randomtest.BcryptHashMatch(),
					),
				},
			},
		},
	})
}

func TestAccResourcePassword_Override(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "override" {
							length = 4
							override_special = "!"
							lower = false
							upper = false
							numeric = false
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.override", tfjsonpath.New("result"), randomtest.StringLengthExact(4)),
					statecheck.ExpectKnownValue("random_password.override", tfjsonpath.New("result"), knownvalue.StringExact("!!!!")),
				},
			},
		},
	})
}

// TestAccResourcePassword_OverrideSpecial_FromVersion3_3_2 verifies behaviour
// when upgrading the provider version from 3.3.2, which set the
// override_special value to null and should not result in a plan difference.
// Reference: https://github.com/hashicorp/terraform-provider-random/issues/306
func TestAccResourcePassword_OverrideSpecial_FromVersion3_3_2(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("override_special"), knownvalue.Null()),
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("override_special"), knownvalue.Null()),
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
				},
			},
		},
	})
}

// TestAccResourcePassword_OverrideSpecial_FromVersion3_4_2 verifies behaviour
// when upgrading the provider version from 3.4.2, which set the
// override_special value to "", while other versions do not.
// Reference: https://github.com/hashicorp/terraform-provider-random/issues/306
func TestAccResourcePassword_OverrideSpecial_FromVersion3_4_2(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion342(),
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("override_special"), knownvalue.StringExact("")),
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("override_special"), knownvalue.Null()),
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
				},
			},
		},
	})
}

func TestAccResourcePassword_ImportWithoutKeepersProducesNoPlannedChanges(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				ResourceName:       "random_password.test",
				ImportStateId:      "Z=:cbrJE?Ltg",
				ImportState:        true,
				ImportStatePersist: true,
			},
			{
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				PlanOnly: true,
			},
		},
	})
}

// TestAccResourcePassword_Import_FromVersion3_1_3 verifies behaviour when resource has been imported and stores
// null for length, lower, number, special, upper, min_lower, min_numeric, min_special, min_upper attributes in state.
// v3.1.3 was selected as this is the last provider version using schema version 0.
func TestAccResourcePassword_Import_FromVersion3_1_3(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion313(),
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				ResourceName:       "random_password.test",
				ImportState:        true,
				ImportStateId:      "Z=:cbrJE?Ltg",
				ImportStatePersist: true,
				// TODO: Import state checks haven't been implemented in terraform-plugin-testing yet, so can't use value comparers for now
				// TODO: Create import state check issue in terraform-plugin-testing
				ImportStateCheck: composeImportStateCheck(
					testCheckNoResourceAttrInstanceState("length"),
					testCheckNoResourceAttrInstanceState("number"),
					testCheckNoResourceAttrInstanceState("upper"),
					testCheckNoResourceAttrInstanceState("lower"),
					testCheckNoResourceAttrInstanceState("special"),
					testCheckNoResourceAttrInstanceState("min_numeric"),
					testCheckNoResourceAttrInstanceState("min_upper"),
					testCheckNoResourceAttrInstanceState("min_lower"),
					testCheckNoResourceAttrInstanceState("min_special"),
					testExtractResourceAttrInstanceState("result", &result1),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("result"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("number"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("numeric"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("upper"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("lower"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("special"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("min_numeric"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("min_upper"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("min_lower"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("min_special"), knownvalue.Int64Exact(0)),
				},
				// TODO: Import state checks haven't been implemented in terraform-plugin-testing yet, so can't use value comparers for now
				// TODO: Create import state check issue in terraform-plugin-testing
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
				),
			},
		},
	})
}

// TestAccResourcePassword_Import_FromVersion3_2_0 verifies behaviour when resource has been imported and stores
// null for length, lower, number, special, upper, min_lower, min_numeric, min_special, min_upper attributes in state.
// v3.2.0 was selected as this is the last provider version using schema version 1.
func TestAccResourcePassword_Import_FromVersion3_2_0(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion320(),
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				ResourceName:       "random_password.test",
				ImportState:        true,
				ImportStateId:      "Z=:cbrJE?Ltg",
				ImportStatePersist: true,
				ImportStateCheck: composeImportStateCheck(
					testCheckNoResourceAttrInstanceState("length"),
					testCheckNoResourceAttrInstanceState("number"),
					testCheckNoResourceAttrInstanceState("upper"),
					testCheckNoResourceAttrInstanceState("lower"),
					testCheckNoResourceAttrInstanceState("special"),
					testCheckNoResourceAttrInstanceState("min_numeric"),
					testCheckNoResourceAttrInstanceState("min_upper"),
					testCheckNoResourceAttrInstanceState("min_lower"),
					testCheckNoResourceAttrInstanceState("min_special"),
					testExtractResourceAttrInstanceState("result", &result1),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				PlanOnly: true,
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("result"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("number"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("numeric"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("upper"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("lower"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("special"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("min_numeric"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("min_upper"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("min_lower"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("min_special"), knownvalue.Int64Exact(0)),
				},
				// TODO: Import state checks haven't been implemented in terraform-plugin-testing yet, so can't use value comparers for now
				// TODO: Create import state check issue in terraform-plugin-testing
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
				),
			},
		},
	})
}

// TestAccResourcePassword_Import_FromVersion3_4_2 verifies behaviour when resource has been imported and stores
// empty map {} for keepers and empty string for override_special in state.
// v3.4.2 was selected as this is the last provider version using schema version 2.
func TestAccResourcePassword_Import_FromVersion3_4_2(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion342(),
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				ResourceName:       "random_password.test",
				ImportState:        true,
				ImportStateId:      "Z=:cbrJE?Ltg",
				ImportStatePersist: true,
				ImportStateCheck: composeImportStateCheck(
					testCheckResourceAttrInstanceState("length"),
					testCheckResourceAttrInstanceState("number"),
					testCheckResourceAttrInstanceState("numeric"),
					testCheckResourceAttrInstanceState("upper"),
					testCheckResourceAttrInstanceState("lower"),
					testCheckResourceAttrInstanceState("special"),
					testCheckResourceAttrInstanceState("min_numeric"),
					testCheckResourceAttrInstanceState("min_upper"),
					testCheckResourceAttrInstanceState("min_lower"),
					testCheckResourceAttrInstanceState("min_special"),
					testExtractResourceAttrInstanceState("result", &result1),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("result"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("number"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("numeric"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("upper"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("lower"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("special"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("min_numeric"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("min_upper"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("min_lower"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("min_special"), knownvalue.Int64Exact(0)),
				},
				// TODO: Import state checks haven't been implemented in terraform-plugin-testing yet, so can't use value comparers for now
				// TODO: Create import state check issue in terraform-plugin-testing
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
				),
			},
		},
	})
}

// TestAccResourcePassword_StateUpgradeV0toV3 covers the state upgrades from V0 to V3.
// This includes the addition of `numeric` and `bcrypt_hash` attributes.
// v3.1.3 is used as this is last version before `bcrypt_hash` attributed was added.
func TestAccResourcePassword_StateUpgradeV0toV3(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                     string
		configBeforeUpgrade      string
		configDuringUpgrade      string
		beforeUpgradeStateChecks []statecheck.StateCheck
		afterUpgradeStateChecks  []statecheck.StateCheck
	}{
		{
			name: "bcrypt_hash",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("bcrypt_hash")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("bcrypt_hash"), knownvalue.NotNull()),
			},
		},
		{
			name: "number is absent before number and numeric are absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is absent before numeric is true during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = true
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is absent before numeric is false during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = false
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(false)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is true before number and numeric are absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is true before numeric is false during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = false
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(false)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is false before numeric is false during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = false
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(false)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(false)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is false before numeric is absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(false)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is false before numeric is true during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = true
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(false)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{"random": {
							VersionConstraint: "3.1.3",
							Source:            "hashicorp/random",
						}},
						Config:            c.configBeforeUpgrade,
						ConfigStateChecks: c.beforeUpgradeStateChecks,
					},
					{
						ProtoV5ProviderFactories: protoV5ProviderFactories(),
						Config:                   c.configDuringUpgrade,
						ConfigStateChecks:        c.afterUpgradeStateChecks,
					},
				},
			})
		})
	}
}

// TestAccResourcePassword_StateUpgrade_V1toV3 covers the state upgrades from V1 to V3.
// This includes the addition of `numeric` attribute.
// v3.2.0 was used as this is the last version before `number` was deprecated and `numeric` attribute
// was added.
func TestAccResourcePassword_StateUpgradeV1toV3(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                     string
		configBeforeUpgrade      string
		configDuringUpgrade      string
		beforeUpgradeStateChecks []statecheck.StateCheck
		afterUpgradeStateChecks  []statecheck.StateCheck
		beforeStateUpgrade       []resource.TestCheckFunc
		afterStateUpgrade        []resource.TestCheckFunc
	}{
		{
			name: "number is absent before number and numeric are absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is absent before numeric is true during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = true
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is absent before numeric is false during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = false
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(false)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is true before numeric is true during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = true
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is true before number and numeric are absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is true before numeric is false during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = false
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(false)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is false before numeric is false during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = false
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(false)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(false)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is false before numeric is absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(false)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
		{
			name: "number is false before numeric is true during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = true
					}`,
			beforeUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(false)),
				randomtest.ExpectNoAttribute("random_password.default", tfjsonpath.New("numeric")),
			},
			afterUpgradeStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("random_password.default", tfjsonpath.New("number"), knownvalue.Bool(true)),
				statecheck.CompareValuePairs(
					"random_password.default", tfjsonpath.New("number"),
					"random_password.default", tfjsonpath.New("numeric"),
					compare.ValuesSame(),
				),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.configDuringUpgrade == "" {
				c.configDuringUpgrade = c.configBeforeUpgrade
			}

			resource.Test(t, resource.TestCase{
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{"random": {
							VersionConstraint: "3.2.0",
							Source:            "hashicorp/random",
						}},
						Config:            c.configBeforeUpgrade,
						ConfigStateChecks: c.beforeUpgradeStateChecks,
					},
					{
						ProtoV5ProviderFactories: protoV5ProviderFactories(),
						Config:                   c.configDuringUpgrade,
						ConfigStateChecks:        c.afterUpgradeStateChecks,
					},
				},
			})
		})
	}
}

func TestAccResourcePassword_Min(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([a-z].*){2,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([A-Z].*){3,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([0-9].*){4,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([!#@])`))),
				},
			},
		},
	})
}

// TestAccResourcePassword_UpgradeFromVersion2_2_1 verifies behaviour when upgrading state from schema V0 to V3.
func TestAccResourcePassword_UpgradeFromVersion2_2_1(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
				t.Skip("This test requires darwin/amd64 to download the old provider version. Setting TF_ACC_TERRAFORM_PATH to darwin/amd64 compatible Terraform binary can be used as a workaround.")
			}
		},
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion221(),
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([a-z].*){2,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([A-Z].*){3,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([0-9].*){4,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([!#@])`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("special"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("upper"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("lower"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("number"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_special"), knownvalue.Int64Exact(1)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_upper"), knownvalue.Int64Exact(3)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_lower"), knownvalue.Int64Exact(2)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_numeric"), knownvalue.Int64Exact(4)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([a-z].*){2,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([A-Z].*){3,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([0-9].*){4,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([!#@])`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("special"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("upper"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("lower"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("number"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("numeric"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_special"), knownvalue.Int64Exact(1)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_upper"), knownvalue.Int64Exact(3)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_lower"), knownvalue.Int64Exact(2)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_numeric"), knownvalue.Int64Exact(4)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("bcrypt_hash"), knownvalue.NotNull()),
				},
			},
		},
	})
}

// TestAccResourcePassword_UpgradeFromVersion3_2_0 verifies behaviour when upgrading state from schema V1 to V3.
func TestAccResourcePassword_UpgradeFromVersion3_2_0(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion320(),
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([a-z].*){2,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([A-Z].*){3,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([0-9].*){4,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([!#@])`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("special"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("upper"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("lower"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("number"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_special"), knownvalue.Int64Exact(1)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_upper"), knownvalue.Int64Exact(3)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_lower"), knownvalue.Int64Exact(2)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_numeric"), knownvalue.Int64Exact(4)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("bcrypt_hash"), knownvalue.NotNull()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				PlanOnly: true,
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([a-z].*){2,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([A-Z].*){3,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([0-9].*){4,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([!#@])`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("special"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("upper"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("lower"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("number"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("numeric"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_special"), knownvalue.Int64Exact(1)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_upper"), knownvalue.Int64Exact(3)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_lower"), knownvalue.Int64Exact(2)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_numeric"), knownvalue.Int64Exact(4)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("bcrypt_hash"), knownvalue.NotNull()),
				},
			},
		},
	})
}

// TestAccResourcePassword_UpgradeFromVersion3_3_2 verifies behaviour when upgrading from SDKv2 to the Framework.
func TestAccResourcePassword_UpgradeFromVersion3_3_2(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([a-z].*){2,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([A-Z].*){3,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([0-9].*){4,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([!#@])`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("special"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("upper"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("lower"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("number"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("numeric"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_special"), knownvalue.Int64Exact(1)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_upper"), knownvalue.Int64Exact(3)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_lower"), knownvalue.Int64Exact(2)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_numeric"), knownvalue.Int64Exact(4)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("bcrypt_hash"), knownvalue.NotNull()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				PlanOnly: true,
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([a-z].*){2,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([A-Z].*){3,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([0-9].*){4,}`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("result"), knownvalue.StringRegexp(regexp.MustCompile(`([!#@])`))),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("special"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("upper"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("lower"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("number"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("numeric"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_special"), knownvalue.Int64Exact(1)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_upper"), knownvalue.Int64Exact(3)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_lower"), knownvalue.Int64Exact(2)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("min_numeric"), knownvalue.Int64Exact(4)),
					statecheck.ExpectKnownValue("random_password.min", tfjsonpath.New("bcrypt_hash"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestUpgradePasswordStateV0toV3(t *testing.T) {
	t.Parallel()

	req := res.UpgradeStateRequest{
		State: &tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id":               tftypes.String,
					"keepers":          tftypes.Map{ElementType: tftypes.String},
					"length":           tftypes.Number,
					"lower":            tftypes.Bool,
					"min_lower":        tftypes.Number,
					"min_numeric":      tftypes.Number,
					"min_special":      tftypes.Number,
					"min_upper":        tftypes.Number,
					"number":           tftypes.Bool,
					"override_special": tftypes.String,
					"result":           tftypes.String,
					"special":          tftypes.Bool,
					"upper":            tftypes.Bool,
				},
			}, map[string]tftypes.Value{
				"id":               tftypes.NewValue(tftypes.String, "none"),
				"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				"length":           tftypes.NewValue(tftypes.Number, 16),
				"lower":            tftypes.NewValue(tftypes.Bool, true),
				"min_lower":        tftypes.NewValue(tftypes.Number, 0),
				"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
				"min_special":      tftypes.NewValue(tftypes.Number, 0),
				"min_upper":        tftypes.NewValue(tftypes.Number, 0),
				"number":           tftypes.NewValue(tftypes.Bool, true),
				"override_special": tftypes.NewValue(tftypes.String, "!#$%\u0026*()-_=+[]{}\u003c\u003e:?"),
				"result":           tftypes.NewValue(tftypes.String, "DZy_3*tnonj%Q%Yx"),
				"special":          tftypes.NewValue(tftypes.Bool, true),
				"upper":            tftypes.NewValue(tftypes.Bool, true),
			}),
			Schema: passwordSchemaV0(),
		},
	}

	resp := &res.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: passwordSchemaV3(),
		},
	}

	upgradePasswordStateV0toV3(context.Background(), req, resp)

	expectedResp := &res.UpgradeStateResponse{
		State: tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bcrypt_hash":      tftypes.String,
					"id":               tftypes.String,
					"keepers":          tftypes.Map{ElementType: tftypes.String},
					"length":           tftypes.Number,
					"lower":            tftypes.Bool,
					"min_lower":        tftypes.Number,
					"min_numeric":      tftypes.Number,
					"min_special":      tftypes.Number,
					"min_upper":        tftypes.Number,
					"number":           tftypes.Bool,
					"numeric":          tftypes.Bool,
					"override_special": tftypes.String,
					"result":           tftypes.String,
					"special":          tftypes.Bool,
					"upper":            tftypes.Bool,
				},
			}, map[string]tftypes.Value{
				"bcrypt_hash":      tftypes.NewValue(tftypes.String, "hash"),
				"id":               tftypes.NewValue(tftypes.String, "none"),
				"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				"length":           tftypes.NewValue(tftypes.Number, 16),
				"lower":            tftypes.NewValue(tftypes.Bool, true),
				"min_lower":        tftypes.NewValue(tftypes.Number, 0),
				"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
				"min_special":      tftypes.NewValue(tftypes.Number, 0),
				"min_upper":        tftypes.NewValue(tftypes.Number, 0),
				"number":           tftypes.NewValue(tftypes.Bool, true),
				"numeric":          tftypes.NewValue(tftypes.Bool, true),
				"override_special": tftypes.NewValue(tftypes.String, "!#$%\u0026*()-_=+[]{}\u003c\u003e:?"),
				"result":           tftypes.NewValue(tftypes.String, "DZy_3*tnonj%Q%Yx"),
				"special":          tftypes.NewValue(tftypes.Bool, true),
				"upper":            tftypes.NewValue(tftypes.Bool, true),
			}),
			Schema: passwordSchemaV3(),
		},
	}

	var bcryptHash, result string

	diags := resp.State.GetAttribute(context.Background(), path.Root("bcrypt_hash"), &bcryptHash)
	if diags.HasError() {
		t.Errorf("error retrieving bcyrpt_hash from state: %s", diags.Errors())
	}

	diags = resp.State.GetAttribute(context.Background(), path.Root("result"), &result)
	if diags.HasError() {
		t.Errorf("error retrieving bcyrpt_hash from state: %s", diags.Errors())
	}

	err := bcrypt.CompareHashAndPassword([]byte(bcryptHash), []byte(result))
	if err != nil {
		t.Errorf("unexpected bcrypt comparison error: %s", err)
	}

	// rawTransformed allows equality testing to be used by mutating the bcrypt_hash value in the response to a known value.
	rawTransformed, err := tftypes.Transform(resp.State.Raw, func(path *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
		bcryptHashPath := tftypes.NewAttributePath().WithAttributeName("bcrypt_hash")

		if path.Equal(bcryptHashPath) {
			return tftypes.NewValue(tftypes.String, "hash"), nil
		}
		return value, nil
	})
	if err != nil {
		t.Errorf("error transforming actual response: %s", err)
	}

	resp.State.Raw = rawTransformed
	if !cmp.Equal(expectedResp, resp) {
		t.Errorf("expected: %+v, got: %+v", expectedResp, resp)
	}
}

func TestUpgradePasswordStateV0toV3_NullValues(t *testing.T) {
	t.Parallel()

	req := res.UpgradeStateRequest{
		State: &tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id":               tftypes.String,
					"keepers":          tftypes.Map{ElementType: tftypes.String},
					"length":           tftypes.Number,
					"lower":            tftypes.Bool,
					"min_lower":        tftypes.Number,
					"min_numeric":      tftypes.Number,
					"min_special":      tftypes.Number,
					"min_upper":        tftypes.Number,
					"number":           tftypes.Bool,
					"override_special": tftypes.String,
					"result":           tftypes.String,
					"special":          tftypes.Bool,
					"upper":            tftypes.Bool,
				},
			}, map[string]tftypes.Value{
				"id":               tftypes.NewValue(tftypes.String, "none"),
				"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				"length":           tftypes.NewValue(tftypes.Number, nil),
				"lower":            tftypes.NewValue(tftypes.Bool, nil),
				"min_lower":        tftypes.NewValue(tftypes.Number, nil),
				"min_numeric":      tftypes.NewValue(tftypes.Number, nil),
				"min_special":      tftypes.NewValue(tftypes.Number, nil),
				"min_upper":        tftypes.NewValue(tftypes.Number, nil),
				"number":           tftypes.NewValue(tftypes.Bool, nil),
				"override_special": tftypes.NewValue(tftypes.String, nil),
				"result":           tftypes.NewValue(tftypes.String, "DZy_3*tnonj%Q%Yx"),
				"special":          tftypes.NewValue(tftypes.Bool, nil),
				"upper":            tftypes.NewValue(tftypes.Bool, nil),
			}),
			Schema: passwordSchemaV0(),
		},
	}

	resp := &res.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: passwordSchemaV3(),
		},
	}

	upgradePasswordStateV0toV3(context.Background(), req, resp)

	expectedResp := &res.UpgradeStateResponse{
		State: tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bcrypt_hash":      tftypes.String,
					"id":               tftypes.String,
					"keepers":          tftypes.Map{ElementType: tftypes.String},
					"length":           tftypes.Number,
					"lower":            tftypes.Bool,
					"min_lower":        tftypes.Number,
					"min_numeric":      tftypes.Number,
					"min_special":      tftypes.Number,
					"min_upper":        tftypes.Number,
					"number":           tftypes.Bool,
					"numeric":          tftypes.Bool,
					"override_special": tftypes.String,
					"result":           tftypes.String,
					"special":          tftypes.Bool,
					"upper":            tftypes.Bool,
				},
			}, map[string]tftypes.Value{
				"bcrypt_hash":      tftypes.NewValue(tftypes.String, "hash"),
				"id":               tftypes.NewValue(tftypes.String, "none"),
				"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				"length":           tftypes.NewValue(tftypes.Number, 16),
				"lower":            tftypes.NewValue(tftypes.Bool, true),
				"min_lower":        tftypes.NewValue(tftypes.Number, 0),
				"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
				"min_special":      tftypes.NewValue(tftypes.Number, 0),
				"min_upper":        tftypes.NewValue(tftypes.Number, 0),
				"number":           tftypes.NewValue(tftypes.Bool, true),
				"numeric":          tftypes.NewValue(tftypes.Bool, true),
				"override_special": tftypes.NewValue(tftypes.String, nil),
				"result":           tftypes.NewValue(tftypes.String, "DZy_3*tnonj%Q%Yx"),
				"special":          tftypes.NewValue(tftypes.Bool, true),
				"upper":            tftypes.NewValue(tftypes.Bool, true),
			}),
			Schema: passwordSchemaV3(),
		},
	}

	var bcryptHash, result string

	diags := resp.State.GetAttribute(context.Background(), path.Root("bcrypt_hash"), &bcryptHash)
	if diags.HasError() {
		t.Errorf("error retrieving bcyrpt_hash from state: %s", diags.Errors())
	}

	diags = resp.State.GetAttribute(context.Background(), path.Root("result"), &result)
	if diags.HasError() {
		t.Errorf("error retrieving bcyrpt_hash from state: %s", diags.Errors())
	}

	err := bcrypt.CompareHashAndPassword([]byte(bcryptHash), []byte(result))
	if err != nil {
		t.Errorf("unexpected bcrypt comparison error: %s", err)
	}

	// rawTransformed allows equality testing to be used by mutating the bcrypt_hash value in the response to a known value.
	rawTransformed, err := tftypes.Transform(resp.State.Raw, func(path *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
		bcryptHashPath := tftypes.NewAttributePath().WithAttributeName("bcrypt_hash")

		if path.Equal(bcryptHashPath) {
			return tftypes.NewValue(tftypes.String, "hash"), nil
		}
		return value, nil
	})
	if err != nil {
		t.Errorf("error transforming actual response: %s", err)
	}

	resp.State.Raw = rawTransformed

	if !cmp.Equal(expectedResp, resp) {
		t.Errorf("expected: %+v, got: %+v", expectedResp, resp)
	}
}

func TestUpgradePasswordStateV1toV3(t *testing.T) {
	t.Parallel()

	req := res.UpgradeStateRequest{
		State: &tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id":               tftypes.String,
					"keepers":          tftypes.Map{ElementType: tftypes.String},
					"length":           tftypes.Number,
					"lower":            tftypes.Bool,
					"min_lower":        tftypes.Number,
					"min_numeric":      tftypes.Number,
					"min_special":      tftypes.Number,
					"min_upper":        tftypes.Number,
					"number":           tftypes.Bool,
					"override_special": tftypes.String,
					"result":           tftypes.String,
					"special":          tftypes.Bool,
					"upper":            tftypes.Bool,
					"bcrypt_hash":      tftypes.String,
				},
			}, map[string]tftypes.Value{
				"id":               tftypes.NewValue(tftypes.String, "none"),
				"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				"length":           tftypes.NewValue(tftypes.Number, 16),
				"lower":            tftypes.NewValue(tftypes.Bool, true),
				"min_lower":        tftypes.NewValue(tftypes.Number, 0),
				"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
				"min_special":      tftypes.NewValue(tftypes.Number, 0),
				"min_upper":        tftypes.NewValue(tftypes.Number, 0),
				"number":           tftypes.NewValue(tftypes.Bool, true),
				"override_special": tftypes.NewValue(tftypes.String, "!#$%\u0026*()-_=+[]{}\u003c\u003e:?"),
				"result":           tftypes.NewValue(tftypes.String, "DZy_3*tnonj%Q%Yx"),
				"special":          tftypes.NewValue(tftypes.Bool, true),
				"upper":            tftypes.NewValue(tftypes.Bool, true),
				"bcrypt_hash":      tftypes.NewValue(tftypes.String, "bcrypt_hash"),
			}),
			Schema: passwordSchemaV1(),
		},
	}

	resp := &res.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: passwordSchemaV3(),
		},
	}

	upgradePasswordStateV1toV3(context.Background(), req, resp)

	expectedResp := &res.UpgradeStateResponse{
		State: tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id":               tftypes.String,
					"keepers":          tftypes.Map{ElementType: tftypes.String},
					"length":           tftypes.Number,
					"lower":            tftypes.Bool,
					"min_lower":        tftypes.Number,
					"min_numeric":      tftypes.Number,
					"min_special":      tftypes.Number,
					"min_upper":        tftypes.Number,
					"number":           tftypes.Bool,
					"numeric":          tftypes.Bool,
					"override_special": tftypes.String,
					"result":           tftypes.String,
					"special":          tftypes.Bool,
					"upper":            tftypes.Bool,
					"bcrypt_hash":      tftypes.String,
				},
			}, map[string]tftypes.Value{
				"id":               tftypes.NewValue(tftypes.String, "none"),
				"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				"length":           tftypes.NewValue(tftypes.Number, 16),
				"lower":            tftypes.NewValue(tftypes.Bool, true),
				"min_lower":        tftypes.NewValue(tftypes.Number, 0),
				"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
				"min_special":      tftypes.NewValue(tftypes.Number, 0),
				"min_upper":        tftypes.NewValue(tftypes.Number, 0),
				"number":           tftypes.NewValue(tftypes.Bool, true),
				"numeric":          tftypes.NewValue(tftypes.Bool, true),
				"override_special": tftypes.NewValue(tftypes.String, "!#$%\u0026*()-_=+[]{}\u003c\u003e:?"),
				"result":           tftypes.NewValue(tftypes.String, "DZy_3*tnonj%Q%Yx"),
				"special":          tftypes.NewValue(tftypes.Bool, true),
				"upper":            tftypes.NewValue(tftypes.Bool, true),
				"bcrypt_hash":      tftypes.NewValue(tftypes.String, "bcrypt_hash"),
			}),
			Schema: passwordSchemaV3(),
		},
	}

	if !cmp.Equal(expectedResp, resp) {
		t.Errorf("expected: %+v, got: %+v", expectedResp, resp)
	}
}

func TestUpgradePasswordStateV1toV3_NullValues(t *testing.T) {
	t.Parallel()

	req := res.UpgradeStateRequest{
		State: &tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id":               tftypes.String,
					"keepers":          tftypes.Map{ElementType: tftypes.String},
					"length":           tftypes.Number,
					"lower":            tftypes.Bool,
					"min_lower":        tftypes.Number,
					"min_numeric":      tftypes.Number,
					"min_special":      tftypes.Number,
					"min_upper":        tftypes.Number,
					"number":           tftypes.Bool,
					"override_special": tftypes.String,
					"result":           tftypes.String,
					"special":          tftypes.Bool,
					"upper":            tftypes.Bool,
					"bcrypt_hash":      tftypes.String,
				},
			}, map[string]tftypes.Value{
				"id":               tftypes.NewValue(tftypes.String, "none"),
				"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				"length":           tftypes.NewValue(tftypes.Number, nil),
				"lower":            tftypes.NewValue(tftypes.Bool, nil),
				"min_lower":        tftypes.NewValue(tftypes.Number, nil),
				"min_numeric":      tftypes.NewValue(tftypes.Number, nil),
				"min_special":      tftypes.NewValue(tftypes.Number, nil),
				"min_upper":        tftypes.NewValue(tftypes.Number, nil),
				"number":           tftypes.NewValue(tftypes.Bool, nil),
				"override_special": tftypes.NewValue(tftypes.String, nil),
				"result":           tftypes.NewValue(tftypes.String, "DZy_3*tnonj%Q%Yx"),
				"special":          tftypes.NewValue(tftypes.Bool, nil),
				"upper":            tftypes.NewValue(tftypes.Bool, nil),
				"bcrypt_hash":      tftypes.NewValue(tftypes.String, "bcrypt_hash"),
			}),
			Schema: passwordSchemaV1(),
		},
	}

	resp := &res.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: passwordSchemaV3(),
		},
	}

	upgradePasswordStateV1toV3(context.Background(), req, resp)

	expectedResp := &res.UpgradeStateResponse{
		State: tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id":               tftypes.String,
					"keepers":          tftypes.Map{ElementType: tftypes.String},
					"length":           tftypes.Number,
					"lower":            tftypes.Bool,
					"min_lower":        tftypes.Number,
					"min_numeric":      tftypes.Number,
					"min_special":      tftypes.Number,
					"min_upper":        tftypes.Number,
					"number":           tftypes.Bool,
					"numeric":          tftypes.Bool,
					"override_special": tftypes.String,
					"result":           tftypes.String,
					"special":          tftypes.Bool,
					"upper":            tftypes.Bool,
					"bcrypt_hash":      tftypes.String,
				},
			}, map[string]tftypes.Value{
				"id":               tftypes.NewValue(tftypes.String, "none"),
				"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				"length":           tftypes.NewValue(tftypes.Number, 16),
				"lower":            tftypes.NewValue(tftypes.Bool, true),
				"min_lower":        tftypes.NewValue(tftypes.Number, 0),
				"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
				"min_special":      tftypes.NewValue(tftypes.Number, 0),
				"min_upper":        tftypes.NewValue(tftypes.Number, 0),
				"number":           tftypes.NewValue(tftypes.Bool, true),
				"numeric":          tftypes.NewValue(tftypes.Bool, true),
				"override_special": tftypes.NewValue(tftypes.String, nil),
				"result":           tftypes.NewValue(tftypes.String, "DZy_3*tnonj%Q%Yx"),
				"special":          tftypes.NewValue(tftypes.Bool, true),
				"upper":            tftypes.NewValue(tftypes.Bool, true),
				"bcrypt_hash":      tftypes.NewValue(tftypes.String, "bcrypt_hash"),
			}),
			Schema: passwordSchemaV3(),
		},
	}

	if !cmp.Equal(expectedResp, resp) {
		t.Errorf("expected: %+v, got: %+v", expectedResp, resp)
	}
}

func TestUpgradePasswordStateV2toV3(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  res.UpgradeStateRequest
		expected *res.UpgradeStateResponse
	}{
		"valid-hash": {
			request: res.UpgradeStateRequest{
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bcrypt_hash":      tftypes.String,
							"id":               tftypes.String,
							"keepers":          tftypes.Map{ElementType: tftypes.String},
							"length":           tftypes.Number,
							"lower":            tftypes.Bool,
							"min_lower":        tftypes.Number,
							"min_numeric":      tftypes.Number,
							"min_special":      tftypes.Number,
							"min_upper":        tftypes.Number,
							"number":           tftypes.Bool,
							"numeric":          tftypes.Bool,
							"override_special": tftypes.String,
							"result":           tftypes.String,
							"special":          tftypes.Bool,
							"upper":            tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"bcrypt_hash":      tftypes.NewValue(tftypes.String, "$2a$10$d9zhEkVg.O1jZ6fEIMRlRuu/vMa0/4UIzeK5joaTBhZJlYiIPhWWa"),
						"id":               tftypes.NewValue(tftypes.String, "none"),
						"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"length":           tftypes.NewValue(tftypes.Number, 20),
						"lower":            tftypes.NewValue(tftypes.Bool, true),
						"min_lower":        tftypes.NewValue(tftypes.Number, 0),
						"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
						"min_special":      tftypes.NewValue(tftypes.Number, 0),
						"min_upper":        tftypes.NewValue(tftypes.Number, 0),
						"number":           tftypes.NewValue(tftypes.Bool, true),
						"numeric":          tftypes.NewValue(tftypes.Bool, true),
						"override_special": tftypes.NewValue(tftypes.String, ""),
						"result":           tftypes.NewValue(tftypes.String, "n:um[a9kO&x!L=9og[EM"),
						"special":          tftypes.NewValue(tftypes.Bool, true),
						"upper":            tftypes.NewValue(tftypes.Bool, true),
					}),
					Schema: passwordSchemaV2(),
				},
			},
			expected: &res.UpgradeStateResponse{
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bcrypt_hash":      tftypes.String,
							"id":               tftypes.String,
							"keepers":          tftypes.Map{ElementType: tftypes.String},
							"length":           tftypes.Number,
							"lower":            tftypes.Bool,
							"min_lower":        tftypes.Number,
							"min_numeric":      tftypes.Number,
							"min_special":      tftypes.Number,
							"min_upper":        tftypes.Number,
							"number":           tftypes.Bool,
							"numeric":          tftypes.Bool,
							"override_special": tftypes.String,
							"result":           tftypes.String,
							"special":          tftypes.Bool,
							"upper":            tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						// The difference checking should compare this actual
						// value since it should not be updated.
						"bcrypt_hash":      tftypes.NewValue(tftypes.String, "$2a$10$d9zhEkVg.O1jZ6fEIMRlRuu/vMa0/4UIzeK5joaTBhZJlYiIPhWWa"),
						"id":               tftypes.NewValue(tftypes.String, "none"),
						"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"length":           tftypes.NewValue(tftypes.Number, 20),
						"lower":            tftypes.NewValue(tftypes.Bool, true),
						"min_lower":        tftypes.NewValue(tftypes.Number, 0),
						"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
						"min_special":      tftypes.NewValue(tftypes.Number, 0),
						"min_upper":        tftypes.NewValue(tftypes.Number, 0),
						"number":           tftypes.NewValue(tftypes.Bool, true),
						"numeric":          tftypes.NewValue(tftypes.Bool, true),
						"override_special": tftypes.NewValue(tftypes.String, ""),
						"result":           tftypes.NewValue(tftypes.String, "n:um[a9kO&x!L=9og[EM"),
						"special":          tftypes.NewValue(tftypes.Bool, true),
						"upper":            tftypes.NewValue(tftypes.Bool, true),
					}),
					Schema: passwordSchemaV3(),
				},
			},
		},
		"invalid-hash": {
			request: res.UpgradeStateRequest{
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bcrypt_hash":      tftypes.String,
							"id":               tftypes.String,
							"keepers":          tftypes.Map{ElementType: tftypes.String},
							"length":           tftypes.Number,
							"lower":            tftypes.Bool,
							"min_lower":        tftypes.Number,
							"min_numeric":      tftypes.Number,
							"min_special":      tftypes.Number,
							"min_upper":        tftypes.Number,
							"number":           tftypes.Bool,
							"numeric":          tftypes.Bool,
							"override_special": tftypes.String,
							"result":           tftypes.String,
							"special":          tftypes.Bool,
							"upper":            tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"bcrypt_hash":      tftypes.NewValue(tftypes.String, "$2a$10$bPOZGBpGe4XIgbpVaWNya.dz/HsU1GDLjuEposH2wf.vUO5rA1wXe"),
						"id":               tftypes.NewValue(tftypes.String, "none"),
						"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"length":           tftypes.NewValue(tftypes.Number, 20),
						"lower":            tftypes.NewValue(tftypes.Bool, true),
						"min_lower":        tftypes.NewValue(tftypes.Number, 0),
						"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
						"min_special":      tftypes.NewValue(tftypes.Number, 0),
						"min_upper":        tftypes.NewValue(tftypes.Number, 0),
						"number":           tftypes.NewValue(tftypes.Bool, true),
						"numeric":          tftypes.NewValue(tftypes.Bool, true),
						"override_special": tftypes.NewValue(tftypes.String, ""),
						"result":           tftypes.NewValue(tftypes.String, "$7r>NiN4Z%uAxpU]:DuB"),
						"special":          tftypes.NewValue(tftypes.Bool, true),
						"upper":            tftypes.NewValue(tftypes.Bool, true),
					}),
					Schema: passwordSchemaV2(),
				},
			},
			expected: &res.UpgradeStateResponse{
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bcrypt_hash":      tftypes.String,
							"id":               tftypes.String,
							"keepers":          tftypes.Map{ElementType: tftypes.String},
							"length":           tftypes.Number,
							"lower":            tftypes.Bool,
							"min_lower":        tftypes.Number,
							"min_numeric":      tftypes.Number,
							"min_special":      tftypes.Number,
							"min_upper":        tftypes.Number,
							"number":           tftypes.Bool,
							"numeric":          tftypes.Bool,
							"override_special": tftypes.String,
							"result":           tftypes.String,
							"special":          tftypes.Bool,
							"upper":            tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						// bcrypt_hash is randomly generated, so the difference checking
						// will ignore this value.
						"bcrypt_hash":      tftypes.NewValue(tftypes.String, nil),
						"id":               tftypes.NewValue(tftypes.String, "none"),
						"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"length":           tftypes.NewValue(tftypes.Number, 20),
						"lower":            tftypes.NewValue(tftypes.Bool, true),
						"min_lower":        tftypes.NewValue(tftypes.Number, 0),
						"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
						"min_special":      tftypes.NewValue(tftypes.Number, 0),
						"min_upper":        tftypes.NewValue(tftypes.Number, 0),
						"number":           tftypes.NewValue(tftypes.Bool, true),
						"numeric":          tftypes.NewValue(tftypes.Bool, true),
						"override_special": tftypes.NewValue(tftypes.String, ""),
						"result":           tftypes.NewValue(tftypes.String, "$7r>NiN4Z%uAxpU]:DuB"),
						"special":          tftypes.NewValue(tftypes.Bool, true),
						"upper":            tftypes.NewValue(tftypes.Bool, true),
					}),
					Schema: passwordSchemaV3(),
				},
			},
		},
		"valid-hash-null-values": {
			request: res.UpgradeStateRequest{
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bcrypt_hash":      tftypes.String,
							"id":               tftypes.String,
							"keepers":          tftypes.Map{ElementType: tftypes.String},
							"length":           tftypes.Number,
							"lower":            tftypes.Bool,
							"min_lower":        tftypes.Number,
							"min_numeric":      tftypes.Number,
							"min_special":      tftypes.Number,
							"min_upper":        tftypes.Number,
							"number":           tftypes.Bool,
							"numeric":          tftypes.Bool,
							"override_special": tftypes.String,
							"result":           tftypes.String,
							"special":          tftypes.Bool,
							"upper":            tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"bcrypt_hash":      tftypes.NewValue(tftypes.String, "$2a$10$d9zhEkVg.O1jZ6fEIMRlRuu/vMa0/4UIzeK5joaTBhZJlYiIPhWWa"),
						"id":               tftypes.NewValue(tftypes.String, "none"),
						"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"length":           tftypes.NewValue(tftypes.Number, nil),
						"lower":            tftypes.NewValue(tftypes.Bool, nil),
						"min_lower":        tftypes.NewValue(tftypes.Number, nil),
						"min_numeric":      tftypes.NewValue(tftypes.Number, nil),
						"min_special":      tftypes.NewValue(tftypes.Number, nil),
						"min_upper":        tftypes.NewValue(tftypes.Number, nil),
						"number":           tftypes.NewValue(tftypes.Bool, nil),
						"numeric":          tftypes.NewValue(tftypes.Bool, nil),
						"override_special": tftypes.NewValue(tftypes.String, ""),
						"result":           tftypes.NewValue(tftypes.String, "n:um[a9kO&x!L=9og[EM"),
						"special":          tftypes.NewValue(tftypes.Bool, nil),
						"upper":            tftypes.NewValue(tftypes.Bool, nil),
					}),
					Schema: passwordSchemaV2(),
				},
			},
			expected: &res.UpgradeStateResponse{
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bcrypt_hash":      tftypes.String,
							"id":               tftypes.String,
							"keepers":          tftypes.Map{ElementType: tftypes.String},
							"length":           tftypes.Number,
							"lower":            tftypes.Bool,
							"min_lower":        tftypes.Number,
							"min_numeric":      tftypes.Number,
							"min_special":      tftypes.Number,
							"min_upper":        tftypes.Number,
							"number":           tftypes.Bool,
							"numeric":          tftypes.Bool,
							"override_special": tftypes.String,
							"result":           tftypes.String,
							"special":          tftypes.Bool,
							"upper":            tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						// The difference checking should compare this actual
						// value since it should not be updated.
						"bcrypt_hash":      tftypes.NewValue(tftypes.String, "$2a$10$d9zhEkVg.O1jZ6fEIMRlRuu/vMa0/4UIzeK5joaTBhZJlYiIPhWWa"),
						"id":               tftypes.NewValue(tftypes.String, "none"),
						"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"length":           tftypes.NewValue(tftypes.Number, 20),
						"lower":            tftypes.NewValue(tftypes.Bool, true),
						"min_lower":        tftypes.NewValue(tftypes.Number, 0),
						"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
						"min_special":      tftypes.NewValue(tftypes.Number, 0),
						"min_upper":        tftypes.NewValue(tftypes.Number, 0),
						"number":           tftypes.NewValue(tftypes.Bool, true),
						"numeric":          tftypes.NewValue(tftypes.Bool, true),
						"override_special": tftypes.NewValue(tftypes.String, ""),
						"result":           tftypes.NewValue(tftypes.String, "n:um[a9kO&x!L=9og[EM"),
						"special":          tftypes.NewValue(tftypes.Bool, true),
						"upper":            tftypes.NewValue(tftypes.Bool, true),
					}),
					Schema: passwordSchemaV3(),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := res.UpgradeStateResponse{
				State: tfsdk.State{
					Schema: testCase.expected.State.Schema,
				},
			}

			upgradePasswordStateV2toV3(context.Background(), testCase.request, &got)

			// Since bcrypt_hash is generated, this test is very involved to
			// ensure the test case is set up properly and the generated
			// value is removed to prevent false positive differences.
			var err error
			var requestBcryptHash, requestResult, expectedBcryptHash, gotBcryptHash, gotResult string

			bcryptHashPath := tftypes.NewAttributePath().WithAttributeName("bcrypt_hash")
			resultPath := tftypes.NewAttributePath().WithAttributeName("result")

			requestBcryptHashValue, err := testTftypesValueAtPath(testCase.request.State.Raw, bcryptHashPath)

			if err != nil {
				t.Fatalf("unexpected error getting request bcrypt_hash value: %s", err)
			}

			if err := requestBcryptHashValue.As(&requestBcryptHash); err != nil {
				t.Fatalf("unexpected error converting request bcrypt_hash to string: %s", err)
			}

			requestResultValue, err := testTftypesValueAtPath(testCase.request.State.Raw, resultPath)

			if err != nil {
				t.Fatalf("unexpected error getting request result value: %s", err)
			}

			if err := requestResultValue.As(&requestResult); err != nil {
				t.Fatalf("unexpected error converting request result to string: %s", err)
			}

			expectedBcryptHashValue, err := testTftypesValueAtPath(testCase.expected.State.Raw, bcryptHashPath)

			if err != nil {
				t.Fatalf("unexpected error getting expected bcrypt_hash value: %s", err)
			}

			if err := expectedBcryptHashValue.As(&expectedBcryptHash); err != nil {
				t.Fatalf("unexpected error converting expected bcrypt_hash to string: %s", err)
			}

			gotBcryptHashValue, err := testTftypesValueAtPath(got.State.Raw, bcryptHashPath)

			if err != nil {
				t.Fatalf("unexpected error getting got bcrypt_hash value: %s", err)
			}

			if err := gotBcryptHashValue.As(&gotBcryptHash); err != nil {
				t.Fatalf("unexpected error converting got bcrypt_hash to string: %s", err)
			}

			gotResultValue, err := testTftypesValueAtPath(got.State.Raw, resultPath)

			if err != nil {
				t.Fatalf("unexpected error getting got result value: %s", err)
			}

			if err := gotResultValue.As(&gotResult); err != nil {
				t.Fatalf("unexpected error converting got result to string: %s", err)
			}

			err = bcrypt.CompareHashAndPassword([]byte(requestBcryptHash), []byte(requestResult))

			// If the request bcrypt_hash was valid, it should be in expected
			// and got. Otherwise, it should be regenerated which will be a
			// random value which must be stripped to prevent false positives.
			if err == nil {
				// Ensure the test case is valid.
				if !requestBcryptHashValue.Equal(expectedBcryptHashValue) {
					t.Fatal("expected request bcrypt_hash in expected")
				}

				// Ensure the request bcrypt_hash was not modified.
				if !requestBcryptHashValue.Equal(gotBcryptHashValue) {
					t.Fatal("expected request bcrypt_hash in got")
				}
			} else {
				// If we got a different error than mismatched hash, then the
				// test case might not be valid.
				if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
					t.Fatalf("unexpected request bcrypt_hash error: %s", err)
				}

				// Ensure the test case has null values on both sides as a
				// regenerated bcrypt_hash cannot be equality compared.
				if !expectedBcryptHashValue.IsNull() {
					t.Fatal("expected null bcrypt_hash in expected")
				}

				// Prevent differences from the got bcrypt_path being randomly
				// generated.
				got.State.Raw, err = tftypes.Transform(
					got.State.Raw,
					func(path *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
						// Purposefully set bcrypt_hash value to nil.
						if path.Equal(bcryptHashPath) {
							return tftypes.NewValue(tftypes.String, nil), nil
						}

						return value, nil
					},
				)

				if err != nil {
					t.Fatalf("unexpected error transforming got: %s", err)
				}
			}

			// The got bcrypt_hash should always be valid.
			if err := bcrypt.CompareHashAndPassword([]byte(gotBcryptHash), []byte(gotResult)); err != nil {
				t.Errorf("unexpected error comparing got bcrypt_hash and result: %s", err)
			}

			// Ensure all state values are checked.
			if diff := cmp.Diff(*testCase.expected, got); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAccResourcePassword_NumberNumericErrors(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "number_numeric_differ" {
  							length = 1
							number = false
  							numeric = true
						}`,
				ExpectError: regexp.MustCompile(`.*Number and numeric are both configured with different values`),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_EmptyMap(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_EmptyMapToNullValue(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_NullMap(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_NullMapToNullValue(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_NullValue(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_NullValues(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_Value(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_Values(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_EmptyMapToValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_NullMapToValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_NullValueToValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_ValueToEmptyMap(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_ValueToNullMap(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_ValueToNullValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_ValueToNewValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_FrameworkMigration_NullMapToNullValue(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_FrameworkMigration_NullMapToValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_FrameworkMigration_NullMapToMultipleNullValue(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_FrameworkMigration_NullMapToMultipleValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_FrameworkMigration_NullMapValue(t *testing.T) {
	// The result attribute values should be the same between test steps
	assertResultSame := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultSame.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourcePassword_Keepers_FrameworkMigration_NullMapValueToValue(t *testing.T) {
	// The result attribute values should differ between test steps
	assertResultDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					assertResultDiffer.AddStateValue("random_password.test", tfjsonpath.New("result")),
					statecheck.ExpectKnownValue("random_password.test", tfjsonpath.New("keepers"), knownvalue.MapSizeExact(2)),
				},
			},
		},
	})
}

func TestAccResourcePassword_NumericFalse(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					special = false
					upper = false
					lower = false
					numeric = false
				}`,
				ExpectError: regexp.MustCompile(`At least one attribute out of \[special,upper,lower,numeric\] must be specified`),
			},
		},
	})
}

func TestAccResourcePassword_NumberFalse(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					special = false
					upper = false
					lower = false
					number = false
				}`,
				ExpectError: regexp.MustCompile(`At least one attribute out of \[special,upper,lower,number\] must be specified`),
			},
		},
	})
}

func TestAccResourcePassword_NumericNumberFalse(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					special = false
					upper = false
					lower = false
					numeric = false
					number = false
				}`,
				ExpectError: regexp.MustCompile(`At least one attribute out of \[special,upper,lower,numeric\] must be specified((.|\n)*)At least one attribute out of \[special,upper,lower,number\] must be specified`),
			},
		},
	})
}

func composeImportStateCheck(fs ...resource.ImportStateCheckFunc) resource.ImportStateCheckFunc {
	return func(s []*terraform.InstanceState) error {
		for i, f := range fs {
			if err := f(s); err != nil {
				return fmt.Errorf("check %d/%d error: %s", i+1, len(fs), err)
			}
		}

		return nil
	}
}

//nolint:unparam
func testExtractResourceAttrInstanceState(attributeName string, attributeValue *string) resource.ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		if len(is) != 1 {
			return fmt.Errorf("unexpected number of instance states: %d", len(is))
		}

		s := is[0]

		attrValue, ok := s.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("attribute %s not found in instance state", attributeName)
		}

		*attributeValue = attrValue

		return nil
	}
}

func testCheckNoResourceAttrInstanceState(attributeName string) resource.ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		if len(is) != 1 {
			return fmt.Errorf("unexpected number of instance states: %d", len(is))
		}

		s := is[0]

		_, ok := s.Attributes[attributeName]
		if ok {
			return fmt.Errorf("attribute %s found in instance state", attributeName)
		}

		return nil
	}
}

func testCheckResourceAttrInstanceState(attributeName string) resource.ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		if len(is) != 1 {
			return fmt.Errorf("unexpected number of instance states: %d", len(is))
		}

		s := is[0]

		_, ok := s.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("attribute %s not found in instance state", attributeName)
		}

		return nil
	}
}
