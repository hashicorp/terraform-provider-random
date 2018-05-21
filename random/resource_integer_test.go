package random

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccResourceIntegerBasic(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testRandomIntegerBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceIntegerBasic("random_integer.integer_1"),
				),
			},
			{
				ResourceName:      "random_integer.integer_1",
				ImportState:       true,
				ImportStateId:     "3,1,3,12345",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceIntegerUpdate(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testRandomIntegerBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceIntegerBasic("random_integer.integer_1"),
				),
			},
			{
				Config: testRandomIntegerUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceIntegerUpdate("random_integer.integer_1"),
				),
			},
		},
	})
}

func TestAccResourceIntegerSeedless_to_seeded(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testRandomIntegerSeedless,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceIntegerSeedless("random_integer.integer_1"),
				),
			},
			{
				Config: testRandomIntegerUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceIntegerUpdate("random_integer.integer_1"),
				),
			},
		},
	})
}

func TestAccResourceIntegerSeeded_to_seedless(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testRandomIntegerBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceIntegerBasic("random_integer.integer_1"),
				),
			},
			{
				Config: testRandomIntegerSeedless,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceIntegerSeedless("random_integer.integer_1"),
				),
			},
		},
	})
}

func testAccResourceIntegerBasic(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not found: %s", id)
		}
		result := rs.Primary.Attributes["result"]

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

func testAccResourceIntegerUpdate(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, _ := s.RootModule().Resources[id]
		// if !ok {
		// 	return fmt.Errorf("Not found: %s", id)
		// }
		result := rs.Primary.Attributes["result"]

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		if result == "" {
			return fmt.Errorf("Result not found")
		}

		if result != "2" {
			return fmt.Errorf("Invalid result %s. Seed does not result in correct value", result)
		}
		return nil
	}
}

// testAccResourceIntegerSeedless only checks that some result was returned, and does not validate the value.
func testAccResourceIntegerSeedless(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not found: %s", id)
		}
		result := rs.Primary.Attributes["result"]

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		if result == "" {
			return fmt.Errorf("Result not found")
		}

		return nil
	}
}

const (
	testRandomIntegerBasic = `
resource "random_integer" "integer_1" {
   min  = 1
   max  = 3
   seed = "12345"
}
`

	testRandomIntegerUpdate = `
resource "random_integer" "integer_1" {
   min  = 1
   max  = 3
   seed = "123456"
}
`

	testRandomIntegerSeedless = `
resource "random_integer" "integer_1" {
   min  = 1
   max  = 3
}
`
)
