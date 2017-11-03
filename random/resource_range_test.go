package random

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccResourceRangeBasic(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testRandomRangeBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceRangeBasic("random_range.range_1"),
				),
			},
		},
	})
}

func testAccResourceRangeBasic(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		result := rs.Primary.ID
		if !ok {
			return fmt.Errorf("Not found: %s", id)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		if result == "" {
			return fmt.Errorf("Result not found")
		}

		if result != "3" {
			return fmt.Errorf("Invalid result %s. Seed does not result in correct value", result)
		}
		return nil
	}
}

const (
	testRandomRangeBasic = `
resource "random_range" "range_1" {
   min  = 1
   max  = 3
   seed = "12345"
}
`
)
