package random

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccResourcePassword(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: strings.ReplaceAll(testAccResourceStringConfig, "random_string", "random_password"),
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
			// TODO: for some reason unable to test import of a single resource here, broken out to test below
		},
	})
}
func TestAccResourcePassword_import(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "random_password" "foo" {
	length = 32
}`,
			},
			{
				ResourceName: "random_password.foo",
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
			return fmt.Errorf("custom string length is %d; want %d", got, want)
		}

		hash := rs.Primary.Attributes["bcrypt_hash"]
		if match, err := regexp.MatchString("^\\$2[ayb]\\$.{56}$", hash); !match || err != nil {
			return fmt.Errorf("Hash does not match regex, error: %v", err)
		}

		return nil
	}
}
