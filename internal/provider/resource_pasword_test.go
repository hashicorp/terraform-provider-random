package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourcePassword(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePasswordConfig,
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
			// Import is tested separately because testAccResourceStringConfig contains 4 resources and during the
			// execution of testStepNewImportState the order of
			// [oldResources](https://github.com/hashicorp/terraform-plugin-sdk/blob/main/helper/resource/testing_new_import_state.go#L177)
			// is non-deterministic. The first resource in oldResources always matches because the check for equality
			// always passes (i.e., ID, type and provider all match because for all 4 resources the values are "none",
			// "random_password" and "registry.terraform.io/hashicorp/random", respectively).
			//
		},
	})
}

func TestAccResourcePassword_import(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "random_password" "foo" {
	length = 32
}`,
			},
			{
				ResourceName: "random_password.foo",
				// Usage of ImportStateIdFunc is required as the value passed to the `terraform import` command needs
				// to be the password itself, as the password resource sets ID to "none" and "result" to the password
				// supplied during import.
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					id := "random_password.foo"
					rs, ok := s.RootModule().Resources[id]
					if !ok {
						return "", fmt.Errorf("Not found: %s", id)
					}
					if rs.Primary.ID == "" {
						return "", fmt.Errorf("No ID is set")
					}

					return rs.Primary.Attributes["result"], nil
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"length", "lower", "number", "special", "upper", "min_lower", "min_numeric", "min_special", "min_upper", "override_special"},
			},
		},
	})
}

const testAccResourcePasswordConfig = `
resource "random_password" "foo" {
  length = 12
}

resource "random_password" "bar" {
  length = 32
}

resource "random_password" "three" {
  length = 4
  override_special = "!"
  lower = false
  upper = false
  number = false
}

resource "random_password" "min" {
  length = 12
  override_special = "!#@"
  min_lower = 2
  min_upper = 3
  min_special = 1
  min_numeric = 4
}
`
