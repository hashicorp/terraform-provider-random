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
				ImportStateVerifyIgnore: []string{"length", "lower", "number", "special", "upper", "min_lower", "min_numeric", "min_special", "min_upper", "override_special"},
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

func TestAccResourcePassword_import_simple(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testRandomPasswordImportSimple,
			},
			{
				ResourceName:            "random_password.bladibla",
				ImportStateId:           "password upper=false,length=8,number=false,special=false,keepers={\"bla\":\"dibla\",\"key\":\"value\"}",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"result"},
				ImportStateCheck:        testAccResourcePasswordImportCheck("random_password.bladibla", "password"),
			},
		},
	})
}

func TestAccResourcePassword_import_stronger(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testRandomPasswordImportStronger,
			},
			{
				ResourceName:            "random_password.bladibla_strong",
				ImportStateId:           "{^%^^]!(]&&([{%%&)]!&)][(^^!(&)) length=32,min_special=32,override_special=_!@#$%^&*()[]{}",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"result"},
				ImportStateCheck:        testAccResourcePasswordImportCheck("random_password.bladibla_strong", "{^%^^]!(]&&([{%%&)]!&)][(^^!(&))"),
			},
		},
	})
}

func testAccResourcePasswordImportCheck(id string, expected string) resource.ImportStateCheckFunc {
	return func(instanceSates []*terraform.InstanceState) error {
		result := instanceSates[0]
		if val, ok := result.Attributes["result"]; ok {
			if val != expected {
				return fmt.Errorf("id %s: result %v is not expected. Expecting %v", id, val, expected)
			}
		}
		return nil
	}
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
	testRandomPasswordImportSimple = `
resource "random_password" "bladibla" {
  keepers          = {
	bla = "dibla"
    key = "value"
  }
  length           = 8
  special          = false
  upper            = false
  lower            = true
  number           = false
  min_numeric      = 0
  min_upper        = 0
  min_lower        = 0
  min_special      = 0
  override_special = ""
}`
	testRandomPasswordImportStronger = `
resource "random_password" "bladibla_strong" {
  length           = 32
  special          = true
  upper            = true
  lower            = true
  number           = true
  min_numeric      = 0
  min_upper        = 0
  min_lower        = 0
  min_special      = 32
  override_special = "_!@#$%^&*()[]{}"
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
