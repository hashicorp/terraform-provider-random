package provider_fm

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourcePet_basic(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePetBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePetLength("random_pet.pet_1", "-", 2),
				),
			},
		},
	})
}

func TestAccResourcePet_length(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePetLengthSet,
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePetLength("random_pet.pet_1", "-", 4),
				),
			},
		},
	})
}

func TestAccResourcePet_prefix(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePetPrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePetLength("random_pet.pet_1", "-", 3),
					resource.TestMatchResourceAttr(
						"random_pet.pet_1", "id", regexp.MustCompile("^consul-")),
				),
			},
		},
	})
}

func TestAccResourcePet_separator(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePetSeparator,
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePetLength("random_pet.pet_1", "_", 3),
				),
			},
		},
	})
}

// nolint:unparam
func testAccResourcePetLength(id string, separator string, length int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not found: %s", id)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		petParts := strings.Split(rs.Primary.ID, separator)
		if len(petParts) != length {
			return fmt.Errorf("Length does not match")
		}

		return nil
	}
}

const testAccResourcePetBasic = `
resource "random_pet" "pet_1" {
}
`

const testAccResourcePetLengthSet = `
resource "random_pet" "pet_1" {
  length = 4
}
`
const testAccResourcePetPrefix = `
resource "random_pet" "pet_1" {
  prefix = "consul"
}
`

const testAccResourcePetSeparator = `
resource "random_pet" "pet_1" {
  length = 3
  separator = "_"
}
`
