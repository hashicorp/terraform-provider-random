package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceUUID(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUUIDConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"random_uuid.basic",
						"result",
						regexp.MustCompile(`[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}`),
					),
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

const (
	testAccResourceUUIDConfig = `
resource "random_uuid" "basic" { 
}
`
)
