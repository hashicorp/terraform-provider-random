package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"golang.org/x/crypto/bcrypt"
)

func TestAccResourcePasswordBasic(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "basic" {
							length = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceStringCheck("random_password.basic", &customLens{
						customLen: 12,
					}),
				),
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
				ImportStateVerifyIgnore: []string{"bcrypt_hash", "length", "lower", "number", "numeric", "special", "upper", "min_lower", "min_numeric", "min_special", "min_upper", "override_special"},
			},
		},
	})
}

func TestAccResourcePasswordOverride(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "override" {
							length = 4
							override_special = "!"
							lower = false
							upper = false
							number = false
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceStringCheck("random_password.override", &customLens{
						customLen: 4,
					}),
					patternMatch("random_password.override", "!!!!"),
				),
			},
		},
	})
}

func TestAccResourcePasswordMin(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
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
				Check: resource.ComposeTestCheckFunc(
					testAccResourceStringCheck("random_password.min", &customLens{
						customLen: 12,
					}),
					regexMatch("random_password.min", regexp.MustCompile(`([a-z])`), 2),
					regexMatch("random_password.min", regexp.MustCompile(`([A-Z])`), 3),
					regexMatch("random_password.min", regexp.MustCompile(`([0-9])`), 4),
					regexMatch("random_password.min", regexp.MustCompile(`([!#@])`), 1),
				),
			},
		},
	})
}

func TestAccResourcePassword_UpdateNumberAndNumeric(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "default" {
							length = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_password.default", "number", "true"),
					resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				),
			},
			{
				Config: `resource "random_password" "default" {
							length = 12
							number = false
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_password.default", "number", "false"),
					resource.TestCheckResourceAttr("random_password.default", "numeric", "false"),
				),
			},
			{
				Config: `resource "random_password" "default" {
							length = 12
							numeric = true
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_password.default", "number", "true"),
					resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				),
			},
			{
				Config: `resource "random_password" "default" {
							length = 12
							numeric = false
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_password.default", "number", "false"),
					resource.TestCheckResourceAttr("random_password.default", "numeric", "false"),
				),
			},
			{
				Config: `resource "random_password" "default" {
							length = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_password.default", "number", "true"),
					resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				),
			},
		},
	})
}

// TestAccResourcePassword_StateUpgraders covers the state upgrades from v0 to V2 and V1 to V2.
// This includes the addition of bcrypt_hash and numeric attributes.
func TestAccResourcePassword_StateUpgraders(t *testing.T) {
	t.Parallel()

	v1Cases := []struct {
		name                string
		configBeforeUpgrade string
		configDuringUpgrade string
		beforeStateUpgrade  []resource.TestCheckFunc
		afterStateUpgrade   []resource.TestCheckFunc
	}{
		{
			name: "%s number is absent",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
			},
		},
		{
			name: "%s number is absent then true",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
			},
		},
		{
			name: "%s number is absent then false",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
			},
		},
		{
			name: "%s number is true",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
			},
		},
		{
			name: "%s number is true then absent",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
			},
		},
		{
			name: "%s number is true then false",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
			},
		},
		{
			name: "%s number is false",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
			},
		},
		{
			name: "%s number is false then absent",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
			},
		},
		{
			name: "%s number is false then true",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "number"),
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
			},
		},
	}

	v0Cases := v1Cases
	v0Cases = append(v0Cases, struct {
		name                string
		configBeforeUpgrade string
		configDuringUpgrade string
		beforeStateUpgrade  []resource.TestCheckFunc
		afterStateUpgrade   []resource.TestCheckFunc
	}{
		name: "%s bcrypt_hash",
		configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
		beforeStateUpgrade: []resource.TestCheckFunc{
			resource.TestCheckNoResourceAttr("random_password.default", "bcrypt_hash"),
		},
		afterStateUpgrade: []resource.TestCheckFunc{
			resource.TestCheckResourceAttrSet("random_password.default", "bcrypt_hash"),
		},
	})

	cases := map[string][]struct {
		name                string
		configBeforeUpgrade string
		configDuringUpgrade string
		beforeStateUpgrade  []resource.TestCheckFunc
		afterStateUpgrade   []resource.TestCheckFunc
	}{
		"3.1.3": v0Cases,
		"3.2.0": v1Cases,
	}

	for providerVersion, v := range cases {
		for _, c := range v {
			name := fmt.Sprintf(c.name, providerVersion)
			t.Run(name, func(t *testing.T) {
				if c.configDuringUpgrade == "" {
					c.configDuringUpgrade = c.configBeforeUpgrade
				}

				resource.UnitTest(t, resource.TestCase{
					Steps: []resource.TestStep{
						{
							ExternalProviders: map[string]resource.ExternalProvider{"random": {
								VersionConstraint: providerVersion,
								Source:            "hashicorp/random",
							}},
							Config: c.configBeforeUpgrade,
							Check:  resource.ComposeTestCheckFunc(c.beforeStateUpgrade...),
						},
						{
							ProviderFactories: testAccProviders,
							Config:            c.configDuringUpgrade,
							Check:             resource.ComposeTestCheckFunc(c.afterStateUpgrade...),
						},
					},
				})
			})
		}
	}
}

func TestResourcePasswordStateUpgradeV0(t *testing.T) {
	cases := []struct {
		name            string
		stateV0         map[string]interface{}
		shouldError     bool
		errMsg          string
		expectedStateV1 map[string]interface{}
	}{
		{
			name:        "result is not string",
			stateV0:     map[string]interface{}{"result": 0},
			shouldError: true,
			errMsg:      "resource password state upgrade failed, result is not a string: int",
		},
		{
			name:            "success",
			stateV0:         map[string]interface{}{"result": "abc123"},
			shouldError:     false,
			expectedStateV1: map[string]interface{}{"result": "abc123", "bcrypt_hash": "123"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actualStateV1, err := resourcePasswordStateUpgradeV0(context.Background(), c.stateV0, nil)

			if c.shouldError {
				// Check error msg
				if !cmp.Equal(c.errMsg, err.Error()) {
					t.Errorf("expected: %q, got: %q", c.errMsg, err)
				}
				// Check actualStateV1 is nil
				if !cmp.Equal(c.expectedStateV1, actualStateV1) {
					t.Errorf("expected: %+v, got: %+v", c.expectedStateV1, err)
				}
			} else {
				if err != nil {
					t.Errorf("err should be nil, actual: %v", err)
				}

				// Compare bcrypt_hash with plaintext equivalent to verify match
				bcryptHash := actualStateV1["bcrypt_hash"]
				err := bcrypt.CompareHashAndPassword([]byte(bcryptHash.(string)), []byte(c.stateV0["result"].(string)))
				if err != nil {
					t.Error("hash and password do not match")
				}

				// Delete bcrypt_hash from actualStateV1 and expectedStateV1 so can compare
				delete(actualStateV1, "bcrypt_hash")
				delete(c.expectedStateV1, "bcrypt_hash")
				if !cmp.Equal(actualStateV1, c.expectedStateV1) {
					t.Errorf("expected: %v, got: %v", c.expectedStateV1, actualStateV1)
				}
			}
		})
	}
}

func TestResourcePasswordStateUpgradeV1(t *testing.T) {
	cases := []struct {
		name            string
		stateV1         map[string]interface{}
		shouldError     bool
		errMsg          string
		expectedStateV2 map[string]interface{}
	}{
		{
			name:        "number is not bool",
			stateV1:     map[string]interface{}{"number": 0},
			shouldError: true,
			errMsg:      "resource password state upgrade failed, number is not a boolean: int",
		},
		{
			name:            "success",
			stateV1:         map[string]interface{}{"number": true},
			shouldError:     false,
			expectedStateV2: map[string]interface{}{"number": true, "numeric": true},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actualStateV2, err := resourcePasswordStateUpgradeV1(context.Background(), c.stateV1, nil)

			if c.shouldError {
				if !cmp.Equal(c.errMsg, err.Error()) {
					t.Errorf("expected: %q, got: %q", c.errMsg, err)
				}
				if !cmp.Equal(c.expectedStateV2, actualStateV2) {
					t.Errorf("expected: %+v, got: %+v", c.expectedStateV2, err)
				}
			} else {
				if err != nil {
					t.Errorf("err should be nil, actual: %v", err)
				}

				if !cmp.Equal(actualStateV2, c.expectedStateV2) {
					t.Errorf("expected: %v, got: %v", c.expectedStateV2, actualStateV2)
				}
			}
		})
	}
}
