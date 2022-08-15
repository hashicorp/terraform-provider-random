package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_id" "foo" {
  							byte_length = 4
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_id.foo", "b64_url", testCheckLen(6)),
					resource.TestCheckResourceAttrWith("random_id.foo", "b64_std", testCheckLen(8)),
					resource.TestCheckResourceAttrWith("random_id.foo", "hex", testCheckLen(8)),
					resource.TestCheckResourceAttrWith("random_id.foo", "dec", testCheckMinLen(1)),
				),
			},
			{
				ResourceName:      "random_id.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceID_ImportWithPrefix(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_id" "bar" {
  							byte_length = 4
  							prefix      = "cloud-"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_id.bar", "b64_url", testCheckLen(12)),
					resource.TestCheckResourceAttrWith("random_id.bar", "b64_std", testCheckLen(14)),
					resource.TestCheckResourceAttrWith("random_id.bar", "hex", testCheckLen(14)),
					resource.TestCheckResourceAttrWith("random_id.bar", "dec", testCheckMinLen(1)),
				),
			},
			{
				ResourceName:        "random_id.bar",
				ImportState:         true,
				ImportStateIdPrefix: "cloud-,",
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccResourceID_UpgradeFromVersion3_3_2(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_id" "bar" {
  							byte_length = 4
  							prefix      = "cloud-"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_id.bar", "b64_url", testCheckLen(12)),
					resource.TestCheckResourceAttrWith("random_id.bar", "b64_std", testCheckLen(14)),
					resource.TestCheckResourceAttrWith("random_id.bar", "hex", testCheckLen(14)),
					resource.TestCheckResourceAttrWith("random_id.bar", "dec", testCheckMinLen(1)),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "bar" {
  							byte_length = 4
  							prefix      = "cloud-"
						}`,
				PlanOnly: true,
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_id" "bar" {
  							byte_length = 4
  							prefix      = "cloud-"
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_id.bar", "b64_url", testCheckLen(12)),
					resource.TestCheckResourceAttrWith("random_id.bar", "b64_std", testCheckLen(14)),
					resource.TestCheckResourceAttrWith("random_id.bar", "hex", testCheckLen(14)),
					resource.TestCheckResourceAttrWith("random_id.bar", "dec", testCheckMinLen(1)),
				),
			},
		},
	})
}
