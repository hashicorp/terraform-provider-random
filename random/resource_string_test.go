package random

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

type customLens struct {
	customLen int
}

func TestAccResourceString(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceStringConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceStringCheck("random_string.foo", &customLens{
						customLen: 12,
					}),
					testAccResourceStringCheck("random_string.bar", &customLens{
						customLen: 32,
					}),
					testAccResourceStringCheck("random_string.three", &customLens{
						customLen: 4,
					}),
					patternMatch("random_string.three", "!!!!"),
				),
			},
		},
	})
}

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

const (
	testAccResourceStringConfig = `
resource "random_string" "foo" {
  length = 12
}

resource "random_string" "bar" {
  length = 32
}

resource "random_string" "three" {
  length = 4
  override_special = "!"
  lower = false
  upper = false
  number = false
}

`
)
