package random

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccResourceChoice(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceChoiceConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccResourceChoiceCheck(
						"random_choice.valid_expect_success",
						[]string{"Ann-Sofie", "Heidi", "Victoria"},
					),
				),
			},
		},
	})
}

func testAccResourceChoiceCheck(id string, wants []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not found: %s", id)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		result := rs.Primary.Attributes["result"]

		if !contains(wants, result) {
			return fmt.Errorf("got result: %s want's one of %s", result, wants)
		}

		return nil
	}
}

func contains(expected []string, actual string) bool {
	for _, a := range expected {
		if a == actual {
			return true
		}
	}
	return false
}

const testAccResourceChoiceConfig = `
resource "random_choice" "valid_expect_success" {
  input = ["Ann-Sofie", "Heidi", "Victoria"]
}
`
