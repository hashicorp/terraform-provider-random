package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourcePet_basic(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_pet" "pet_1" {
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_pet.pet_1", "id", testCheckPetLen("-", 2)),
				),
			},
		},
	})
}

func TestAccResourcePet_length(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_pet" "pet_1" {
  							length = 4
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_pet.pet_1", "id", testCheckPetLen("-", 4)),
				),
			},
		},
	})
}

func TestAccResourcePet_prefix(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_pet" "pet_1" {
  							prefix = "consul"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_pet.pet_1", "id", testCheckPetLen("-", 3)),
					resource.TestMatchResourceAttr("random_pet.pet_1", "id", regexp.MustCompile("^consul-")),
				),
			},
		},
	})
}

func TestAccResourcePet_separator(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_pet" "pet_1" {
  							length = 3
  							separator = "_"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_pet.pet_1", "id", testCheckPetLen("_", 3)),
				),
			},
		},
	})
}

func testCheckPetLen(separator string, expectedLen int) func(input string) error {
	return func(input string) error {
		petNameParts := strings.Split(input, separator)

		if len(petNameParts) != expectedLen {
			return fmt.Errorf("expected length %d, actual length %d", expectedLen, len(petNameParts))
		}

		return nil
	}
}
