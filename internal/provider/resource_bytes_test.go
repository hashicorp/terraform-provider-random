package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceBytes(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_bytes" "basic" {
							length = 32
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("random_bytes.basic", "result_base64", regexp.MustCompile(`^[A-Za-z/+\d]{43}=$`)),
					resource.TestMatchResourceAttr("random_bytes.basic", "result_hex", regexp.MustCompile(`^[a-f\d]{64}$`)),
					resource.TestCheckResourceAttr("random_bytes.basic", "length", "32"),
				),
			},
			{
				// Usage of ImportStateIdFunc is required as the value passed to the `terraform import` command needs
				// to be the bytes encoded with base64, as the bytes resource sets ID to "none"
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					id := "random_bytes.basic"
					rs, ok := s.RootModule().Resources[id]
					if !ok {
						return "", fmt.Errorf("not found: %s", id)
					}
					if rs.Primary.ID == "" {
						return "", fmt.Errorf("no ID is set")
					}

					return rs.Primary.Attributes["result_base64"], nil
				},
				ResourceName:      "random_bytes.basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceBytes_LengthErrors(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_bytes" "invalid_length" {
							length = 0
						}`,
				ExpectError: regexp.MustCompile(`.*Attribute length value must be at least 1, got: 0`),
			},
		},
	})
}

func TestAccResourceBytes_Length_ForceReplacement(t *testing.T) {
	var bytes1, bytes2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 1
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_bytes.test", "length", "1"),
					testExtractResourceAttr("random_bytes.test", "result_base64", &bytes1),
					resource.TestCheckResourceAttrWith("random_bytes.test", "result_hex", testCheckLen(2)),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_bytes" "test" {
					length = 2
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("random_bytes.test", "length", "2"),
					testExtractResourceAttr("random_bytes.test", "result_base64", &bytes2),
					resource.TestCheckResourceAttrWith("random_bytes.test", "result_hex", testCheckLen(4)),
					testCheckAttributeValuesDiffer(&bytes1, &bytes2),
				),
			},
		},
	})
}
