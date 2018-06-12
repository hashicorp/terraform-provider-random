package random

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccResourcePassword(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePasswordConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePasswordCheck("random_password.foo", &customLens{
						customLen: 12,
					}),
					testAccResourcePasswordCheck("random_password.bar", &customLens{
						customLen: 32,
					}),
					testAccResourcePasswordCheck("random_password.three", &customLens{
						customLen: 4,
					}),
					patternMatch("random_password.three", "!!!!"),
					testAccResourcePasswordCheck("random_password.min", &customLens{
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

func testAccResourcePasswordCheck(id string, want *customLens) resource.TestCheckFunc {
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
			return fmt.Errorf("custom password length is %d; want %d", got, want)
		}

		return nil
	}
}

const (
	testAccResourcePasswordConfig = `
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
)
