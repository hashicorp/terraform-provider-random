package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceIntegerBasic(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_integer" "integer_1" {
   							min  = 1
							max  = 3
   							seed = "12345"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_integer.integer_1", "result", "3"),
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
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_integer" "integer_1" {
   							min  = 1
							max  = 3
   							seed = "12345"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_integer.integer_1", "result", "3"),
				),
			},
			{
				Config: `resource "random_integer" "integer_1" {
							min  = 1
   							max  = 3
   							seed = "123456"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_integer.integer_1", "result", "2"),
				),
			},
		},
	})
}

func TestAccResourceIntegerSeedless_to_seeded(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_integer" "integer_1" {
   							min  = 1
   							max  = 3
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_integer.integer_1", "result", testCheckNotEmptyString("result")),
				),
			},
			{
				Config: `resource "random_integer" "integer_1" {
							min  = 1
   							max  = 3
   							seed = "123456"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_integer.integer_1", "result", "2"),
				),
			},
		},
	})
}

func TestAccResourceIntegerSeeded_to_seedless(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_integer" "integer_1" {
   							min  = 1
							max  = 3
   							seed = "12345"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_integer.integer_1", "result", "3"),
				),
			},
			{
				Config: `resource "random_integer" "integer_1" {
   							min  = 1
   							max  = 3
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_integer.integer_1", "result", testCheckNotEmptyString("result")),
				),
			},
		},
	})
}

func TestAccResourceIntegerBig(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_integer" "integer_1" {
   							max  = 7227701560655103598
   							min  = 7227701560655103597
   							seed = 12345
						}`,
			},
			{
				ResourceName:      "random_integer.integer_1",
				ImportState:       true,
				ImportStateId:     "7227701560655103598,7227701560655103597,7227701560655103598,12345",
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckNotEmptyString(field string) func(input string) error {
	return func(input string) error {
		if input == "" {
			return fmt.Errorf("%s is empty string", field)
		}

		return nil
	}
}
