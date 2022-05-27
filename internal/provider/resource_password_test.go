package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

func TestAccResourcePassword_V0ToV2(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{"random": {
					VersionConstraint: "3.1.3",
					Source:            "hashicorp/random",
				}},
				Config: `terraform {
							required_providers {
								random = {
									source = "hashicorp/random"
									version = "3.1.3"
								}
							}
						}
						provider "random" {}
						resource "random_password" "default" {
							length = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("random_password.default", "bcrypt_hash"),
					resource.TestCheckResourceAttrSet("random_password.default", "numeric"),
				),
			},
		},
	})
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
				if !cmp.Equal(c.errMsg, err.Error()) {
					t.Errorf("expected: %q, got: %q", c.errMsg, err)
				}
				if !cmp.Equal(c.expectedStateV1, actualStateV1) {
					t.Errorf("expected: %+v, got: %+v", c.expectedStateV1, err)
				}
			} else {
				if err != nil {
					t.Errorf("err should be nil, actual: %v", err)
				}

				for k := range c.expectedStateV1 {
					_, ok := actualStateV1[k]
					if !ok {
						t.Errorf("expected key: %s is missing from state", k)
					}
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

				for k := range c.expectedStateV2 {
					_, ok := actualStateV2[k]
					if !ok {
						t.Errorf("expected key: %s is missing from state", k)
					}
				}
			}
		})
	}
}
