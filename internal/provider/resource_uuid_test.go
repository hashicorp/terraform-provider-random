package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceUUID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUUIDConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceUUIDCheck("random_uuid.basic"),
				),
			},
			{
				ResourceName:      "random_uuid.basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceUUIDCheck(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not found: %s", id)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		result := rs.Primary.Attributes["result"]
		matched, err := regexp.MatchString(
			"[\\da-f]{8}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{12}", result)
		if !matched || err != nil {
			return fmt.Errorf("result string format incorrect, is %s", result)
		}

		return nil
	}
}

const (
	testAccResourceUUIDConfig = `
resource "random_uuid" "basic" { }
`
)
