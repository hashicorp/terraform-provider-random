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

type customLens struct {
	customLen int
}

func TestAccResourceString(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceStringBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceStringCheck("random_string.basic", &customLens{
						customLen: 12,
					}),
				),
			},
			{
				ResourceName:            "random_string.basic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"length", "lower", "number", "numeric", "special", "upper", "min_lower", "min_numeric", "min_special", "min_upper", "override_special"},
			},
		},
	})
}

func TestAccResourceStringOverride(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceStringOverride,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceStringCheck("random_string.override", &customLens{
						customLen: 4,
					}),
					patternMatch("random_string.override", "!!!!"),
				),
			},
		},
	})
}

func TestAccResourceStringMin(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceStringMin,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceStringCheck("random_string.min", &customLens{
						customLen: 12,
					}),
					regexMatch("random_string.min", regexp.MustCompile(`([a-z])`), 2),
					regexMatch("random_string.min", regexp.MustCompile(`([A-Z])`), 3),
					regexMatch("random_string.min", regexp.MustCompile(`([0-9])`), 4),
					regexMatch("random_string.min", regexp.MustCompile(`([!#@])`), 1),
				),
			},
		},
	})
}

func TestAccResourceString_UpdateNumberAndNumeric(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `resource "random_string" "default" {
							length = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_string.default", "number", "true"),
					resource.TestCheckResourceAttr("random_string.default", "numeric", "true"),
				),
			},
			{
				Config: `resource "random_string" "default" {
							length = 12
							number = false
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_string.default", "number", "false"),
					resource.TestCheckResourceAttr("random_string.default", "numeric", "false"),
				),
			},
			{
				Config: `resource "random_string" "default" {
							length = 12
							numeric = true
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_string.default", "number", "true"),
					resource.TestCheckResourceAttr("random_string.default", "numeric", "true"),
				),
			},
			{
				Config: `resource "random_string" "default" {
							length = 12
							numeric = false
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_string.default", "number", "false"),
					resource.TestCheckResourceAttr("random_string.default", "numeric", "false"),
				),
			},
			{
				Config: `resource "random_string" "default" {
							length = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_string.default", "number", "true"),
					resource.TestCheckResourceAttr("random_string.default", "numeric", "true"),
				),
			},
		},
	})
}

// TestAccResourceString_StateUpgraders covers the state upgrade from V1 to V2.
// This includes the addition of numeric attribute.
func TestAccResourceString_StateUpgraders(t *testing.T) {
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
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "%s number is absent then true",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_string" "default" {
						length = 12
						number = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "%s number is absent then false",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_string" "default" {
						length = 12
						number = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "%s number is true",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
						number = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "%s number is true then absent",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_string" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "%s number is true then false",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_string" "default" {
						length = 12
						number = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "%s number is false",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
						number = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "%s number is false then absent",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_string" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "%s number is false then true",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_string" "default" {
						length = 12
						number = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_string.default", "number"),
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
	}

	cases := map[string][]struct {
		name                string
		configBeforeUpgrade string
		configDuringUpgrade string
		beforeStateUpgrade  []resource.TestCheckFunc
		afterStateUpgrade   []resource.TestCheckFunc
	}{
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

func TestResourceStringStateUpgradeV1(t *testing.T) {
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
			errMsg:      "resource string state upgrade failed, number is not a boolean: int",
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
			actualStateV2, err := resourceStringStateUpgradeV1(context.Background(), c.stateV1, nil)

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

func TestAccResourceStringErrors(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceStringInvalidConfig,
				ExpectError: regexp.MustCompile(`.*length \(2\) must be >= min_upper \+ min_lower \+ min_numeric \+ min_special \(3\)`),
			},
			{
				Config:      testAccResourceStringLengthTooShortConfig,
				ExpectError: regexp.MustCompile(`.*expected length to be at least \(1\), got 0`),
			},
		},
	})
}

const (
	testAccResourceStringBasic = `
resource "random_string" "basic" {
  length = 12
}`
	testAccResourceStringOverride = `
resource "random_string" "override" {
length = 4
override_special = "!"
lower = false
upper = false
number = false
}
`
	testAccResourceStringMin = `
resource "random_string" "min" {
length = 12
override_special = "!#@"
min_lower = 2
min_upper = 3
min_special = 1
min_numeric = 4
}`
	testAccResourceStringInvalidConfig = `
resource "random_string" "invalid_length" {
  length = 2
  min_lower = 3
}`
	testAccResourceStringLengthTooShortConfig = `
resource "random_string" "invalid_length" {
  length = 0
}`
)

func testAccResourceStringCheck(id string, want *customLens) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not found: %s", id)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		customStr := rs.Primary.Attributes["result"]

		if got, want := len(customStr), want.customLen; got != want {
			return fmt.Errorf("custom string length is %d; want %d", got, want)
		}

		return nil
	}
}

func regexMatch(id string, exp *regexp.Regexp, requiredMatches int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not found: %s", id)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		customStr := rs.Primary.Attributes["result"]

		if matches := exp.FindAllStringSubmatchIndex(customStr, -1); len(matches) < requiredMatches {
			return fmt.Errorf("custom string is %s; did not match %s", customStr, exp)
		}

		return nil
	}
}
func patternMatch(id string, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not found: %s", id)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		customStr := rs.Primary.Attributes["result"]

		if got, want := customStr, want; got != want {
			return fmt.Errorf("custom string is %s; want %s", got, want)
		}

		return nil
	}
}
