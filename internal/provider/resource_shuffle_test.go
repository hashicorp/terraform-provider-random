package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// These results are current as of Go 1.6. The Go
// "rand" package does not guarantee that the random
// number generator will generate the same results
// forever, but the maintainers endeavor not to change
// it gratuitously.
// These tests allow us to detect such changes and
// document them when they arise, but the docs for this
// resource specifically warn that results are not
// guaranteed consistent across Terraform releases.
func TestAccResourceShuffle(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_shuffle" "default_length" {
    						input = ["a", "b", "c", "d", "e"]
    						seed = "-"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_shuffle.default_length", "result.#", testAccResourceShuffleCheckLength("5")),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.0", "a"),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.1", "c"),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.2", "b"),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.3", "e"),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.4", "d"),
				),
			},
		},
	})
}

func TestAccResourceShuffle_Shorter(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_shuffle" "shorter_length" {
    						input = ["a", "b", "c", "d", "e"]
    						seed = "-"
    						result_count = 3
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_shuffle.shorter_length", "result.#", testAccResourceShuffleCheckLength("3")),
					resource.TestCheckResourceAttr("random_shuffle.shorter_length", "result.0", "a"),
					resource.TestCheckResourceAttr("random_shuffle.shorter_length", "result.1", "c"),
					resource.TestCheckResourceAttr("random_shuffle.shorter_length", "result.2", "b"),
				),
			},
		},
	})
}

func TestAccResourceShuffle_Longer(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_shuffle" "longer_length" {
    						input = ["a", "b", "c", "d", "e"]
    						seed = "-"
    						result_count = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_shuffle.longer_length", "result.#", testAccResourceShuffleCheckLength("12")),
					resource.TestCheckResourceAttr("random_shuffle.longer_length", "result.0", "a"),
					resource.TestCheckResourceAttr("random_shuffle.longer_length", "result.1", "c"),
					resource.TestCheckResourceAttr("random_shuffle.longer_length", "result.2", "b"),
					resource.TestCheckResourceAttr("random_shuffle.longer_length", "result.3", "e"),
					resource.TestCheckResourceAttr("random_shuffle.longer_length", "result.4", "d"),
					resource.TestCheckResourceAttr("random_shuffle.longer_length", "result.5", "a"),
					resource.TestCheckResourceAttr("random_shuffle.longer_length", "result.6", "e"),
					resource.TestCheckResourceAttr("random_shuffle.longer_length", "result.7", "d"),
					resource.TestCheckResourceAttr("random_shuffle.longer_length", "result.8", "c"),
					resource.TestCheckResourceAttr("random_shuffle.longer_length", "result.9", "b"),
					resource.TestCheckResourceAttr("random_shuffle.longer_length", "result.10", "a"),
					resource.TestCheckResourceAttr("random_shuffle.longer_length", "result.11", "b"),
				),
			},
		},
	})
}

func TestAccResourceShuffle_Empty(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_shuffle" "empty_length" {
    						input = []
    						seed = "-"
    						result_count = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_shuffle.empty_length", "result.#", testAccResourceShuffleCheckLength("0")),
				),
			},
		},
	})
}

func TestAccResourceShuffle_One(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_shuffle" "one_length" {
    						input = ["a"]
    						seed = "-"
    						result_count = 1
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_shuffle.one_length", "result.#", testAccResourceShuffleCheckLength("1")),
					resource.TestCheckResourceAttr("random_shuffle.one_length", "result.0", "a"),
				),
			},
		},
	})
}

func TestAccResourceShuffle_UpgradeFromVersion3_3_2(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_shuffle" "default_length" {
    						input = ["a", "b", "c", "d", "e"]
    						seed = "-"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_shuffle.default_length", "result.#", testAccResourceShuffleCheckLength("5")),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.0", "a"),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.1", "c"),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.2", "b"),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.3", "e"),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.4", "d"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "default_length" {
    						input = ["a", "b", "c", "d", "e"]
    						seed = "-"
						}`,
				PlanOnly: true,
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_shuffle" "default_length" {
    						input = ["a", "b", "c", "d", "e"]
    						seed = "-"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_shuffle.default_length", "result.#", testAccResourceShuffleCheckLength("5")),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.0", "a"),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.1", "c"),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.2", "b"),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.3", "e"),
					resource.TestCheckResourceAttr("random_shuffle.default_length", "result.4", "d"),
				),
			},
		},
	})
}

func testAccResourceShuffleCheckLength(expectedLength string) func(input string) error {
	return func(input string) error {
		if input != expectedLength {
			return fmt.Errorf("got length %s; expected length %s", input, expectedLength)
		}

		return nil
	}
}
