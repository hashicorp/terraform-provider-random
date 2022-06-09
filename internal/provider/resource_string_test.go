package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type customLens struct {
	customLen int
}

func TestAccResourceString(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
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

// TestAccResourceString_StateUpgrade_V1toV2 covers the state upgrade from V1 to V2.
// This includes the deprecation of `number` and the addition of `numeric` attributes.
// v3.2.0 was used as this is the last version before `number` was deprecated and `numeric` attribute
// was added.
func TestAccResourceString_StateUpgraders(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		configBeforeUpgrade string
		configDuringUpgrade string
		beforeStateUpgrade  []resource.TestCheckFunc
		afterStateUpgrade   []resource.TestCheckFunc
	}{
		{
			name: "number is absent",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "number is absent then true",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_string" "default" {
						length = 12
						number = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "number is absent then false",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_string" "default" {
						length = 12
						number = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "number is true",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
						number = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "number is true then absent",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_string" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "number is true then false",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_string" "default" {
						length = 12
						number = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "number is false",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
						number = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "number is false then absent",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_string" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
		{
			name: "number is false then true",
			configBeforeUpgrade: `resource "random_string" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_string" "default" {
						length = 12
						number = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_string.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_string.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_string.default", "number", "random_string.default", "numeric"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.configDuringUpgrade == "" {
				c.configDuringUpgrade = c.configBeforeUpgrade
			}

			resource.UnitTest(t, resource.TestCase{
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{"random": {
							VersionConstraint: "3.2.0",
							Source:            "hashicorp/random",
						}},
						Config: c.configBeforeUpgrade,
						Check:  resource.ComposeTestCheckFunc(c.beforeStateUpgrade...),
					},
					{
						ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
						Config:                   c.configDuringUpgrade,
						Check:                    resource.ComposeTestCheckFunc(c.afterStateUpgrade...),
					},
				},
			})
		})
	}
}

func TestAccResourceStringErrors(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceStringInvalidConfig,
				ExpectError: regexp.MustCompile(`.*Attribute "length" \(2\) cannot be less than min_upper \+ min_lower \+\nmin_numeric \+ min_special \(3\).`),
			},
			{
				Config:      testAccResourceStringLengthTooShortConfig,
				ExpectError: regexp.MustCompile(`.*Attribute "length" \(0\) must be at least 1`),
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
