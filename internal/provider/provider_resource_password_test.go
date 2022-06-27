package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourcePasswordBasic(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "basic" {
							length = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_password.basic", "result", testCheckLen(12)),
				),
			},
			{
				ResourceName: "random_password.basic",
				// Usage of ImportStateIdFunc is required as the value passed to the `terraform import` command needs
				// to be the password itself, as the password resource sets ID to "none" and "result" to the password
				// supplied during import.
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					id := "random_password.basic"
					rs, ok := s.RootModule().Resources[id]
					if !ok {
						return "", fmt.Errorf("not found: %s", id)
					}
					if rs.Primary.ID == "" {
						return "", fmt.Errorf("no ID is set")
					}

					return rs.Primary.Attributes["result"], nil
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bcrypt_hash"},
			},
		},
	})
}

func TestAccResourcePasswordOverride(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "override" {
							length = 4
							override_special = "!"
							lower = false
							upper = false
							numeric = false
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_password.override", "result", testCheckLen(4)),
					resource.TestCheckResourceAttr("random_password.override", "result", "!!!!"),
				),
			},
		},
	})
}

// TestAccResourcePassword_StateUpgrade_V0toV2 covers the state upgrades from V0 to V2.
// This includes the deprecation and removal of `number` and the addition of `numeric`
// and `bcrypt_hash` attributes.
// v3.1.3 is used as this is last version before `bcrypt_hash` attributed was added.
func TestAccResourcePassword_StateUpgrade_V0toV2(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		configBeforeUpgrade string
		configDuringUpgrade string
		beforeStateUpgrade  []resource.TestCheckFunc
		afterStateUpgrade   []resource.TestCheckFunc
	}{
		{
			name: "bcrypt_hash",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckNoResourceAttr("random_password.default", "bcrypt_hash"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("random_password.default", "bcrypt_hash"),
			},
		},
		{
			name: "number is absent before numeric is absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is absent before numeric is true during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is absent before numeric is false during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is true before numeric is absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is true before numeric is absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is true before numeric is false during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is false before numeric is false during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is false before numeric is absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is false before numeric is true during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{"random": {
							VersionConstraint: "3.1.3",
							Source:            "hashicorp/random",
						}},
						Config: c.configBeforeUpgrade,
						Check:  resource.ComposeTestCheckFunc(c.beforeStateUpgrade...),
					},
					{
						ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
						Config:                   c.configDuringUpgrade,
						Check:                    resource.ComposeTestCheckFunc(c.afterStateUpgrade...),
					},
				},
			})
		})
	}
}

// TestAccResourcePassword_StateUpgrade_V1toV2 covers the state upgrades from V1 to V2.
// This includes the deprecation and removal of `number` and the addition of `numeric` attributes.
// v3.2.0 was used as this is the last version before `number` was deprecated and `numeric` attribute
// was added.
func TestAccResourcePassword_StateUpgrade_V1toV2(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		configBeforeUpgrade string
		configDuringUpgrade string
		beforeStateUpgrade  []resource.TestCheckFunc
		afterStateUpgrade   []resource.TestCheckFunc
	}{
		{
			name: "number is absent before numeric is absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is absent before numeric is true during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is absent before numeric is false during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is true before numeric is true during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is true before numeric is absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is true before numeric is false during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is false before numeric is false during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is false before numeric is absent during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
		{
			name: "number is false before numeric is true during",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						numeric = true
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "numeric", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "number"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.configDuringUpgrade == "" {
				c.configDuringUpgrade = c.configBeforeUpgrade
			}

			// TODO: Why is resource.Test not being used here
			resource.UnitTest(t, resource.TestCase{
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{"random": {
							VersionConstraint: "3.2.0",
							Source:            "hashicorp/random",
						}},
						Config: c.configBeforeUpgrade,
						Check:  resource.ComposeTestCheckFunc(c.beforeStateUpgrade...),
					},
					{
						ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
						Config:                   c.configDuringUpgrade,
						Check:                    resource.ComposeTestCheckFunc(c.afterStateUpgrade...),
					},
				},
			})
		})
	}
}

func TestAccResourcePasswordMin(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("random_password.min", "result", testCheckLen(12)),
					resource.TestMatchResourceAttr("random_password.min", "result", regexp.MustCompile(`([a-z].*){2,}`)),
					resource.TestMatchResourceAttr("random_password.min", "result", regexp.MustCompile(`([A-Z].*){3,}`)),
					resource.TestMatchResourceAttr("random_password.min", "result", regexp.MustCompile(`([0-9].*){4,}`)),
					resource.TestMatchResourceAttr("random_password.min", "result", regexp.MustCompile(`([!#@])`)),
				),
			},
		},
	})
}
