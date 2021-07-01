package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
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

func TestAccResourcePassword(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: strings.Replace(testAccResourceStringConfig, "random_string", "random_password", -1),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceStringCheck("random_password.foo", &customLens{
						customLen: 12,
					}),
					testAccResourceStringCheck("random_password.bar", &customLens{
						customLen: 32,
					}),
					testAccResourceStringCheck("random_password.three", &customLens{
						customLen: 4,
					}),
					patternMatch("random_password.three", "!!!!"),
					testAccResourceStringCheck("random_password.min", &customLens{
						customLen: 12,
					}),
					regexMatch("random_password.min", regexp.MustCompile(`([a-z])`), 2),
					regexMatch("random_password.min", regexp.MustCompile(`([A-Z])`), 3),
					regexMatch("random_password.min", regexp.MustCompile(`([0-9])`), 4),
					regexMatch("random_password.min", regexp.MustCompile(`([!#@])`), 1),
				),
			},
			// TODO: for some reason unable to test import of a single resource here, broken out to test below
		},
	})
}

func TestAccResourcePassword_import_simple(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
				return fmt.Errorf("result %v is not expected. Expecting %v", val, expected)
			}
		}
		return nil
	}
}
