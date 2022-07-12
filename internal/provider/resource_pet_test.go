package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourcePet(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
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

func TestAccResourcePet_Length(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
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

func TestAccResourcePet_Prefix(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
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

func TestAccResourcePet_Separator(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
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

func TestAccResourcePet_UpgradeFromVersion3_3_2(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_pet" "pet_1" {
  							prefix = "consul"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_pet.pet_1", "id", testCheckPetLen("-", 3)),
					resource.TestMatchResourceAttr("random_pet.pet_1", "id", regexp.MustCompile("^consul-")),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_pet" "pet_1" {
  							prefix = "consul"
						}`,
				PlanOnly: true,
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
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

func testCheckPetLen(separator string, expectedLen int) func(input string) error {
	return func(input string) error {
		petNameParts := strings.Split(input, separator)

		if len(petNameParts) != expectedLen {
			return fmt.Errorf("expected length %d, actual length %d", expectedLen, len(petNameParts))
		}

		return nil
	}
}
