package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	res "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-random/internal/random"
	"golang.org/x/crypto/bcrypt"
)

func TestGenerateHash(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input random.StringParams
	}{
		"defaults": {
			input: random.StringParams{
				Length:  32, // Required
				Lower:   true,
				Numeric: true,
				Special: true,
				Upper:   true,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			randomBytes, err := random.CreateString(testCase.input)

			if err != nil {
				t.Fatalf("unexpected random.CreateString error: %s", err)
			}

			hash, err := generateHash(string(randomBytes))

			if err != nil {
				t.Fatalf("unexpected generateHash error: %s", err)
			}

			err = bcrypt.CompareHashAndPassword([]byte(hash), randomBytes)

			if err != nil {
				t.Fatalf("unexpected bcrypt.CompareHashAndPassword error: %s", err)
			}
		})
	}
}

func TestAccResourcePassword(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
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

func TestAccResourcePassword_BcryptHash(t *testing.T) {
	t.Parallel()

	var result, bcryptHash string

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "bcrypt_hash", &bcryptHash),
					testExtractResourceAttr("random_password.test", "result", &result),
					testBcryptHashValid(&bcryptHash, &result),
				),
			},
		},
	})
}

// TestAccResourcePassword_BcryptHash_FromVersion3_3_2 verifies behaviour when
// upgrading state from schema V2 to V3 without a bcrypt_hash update.
func TestAccResourcePassword_BcryptHash_FromVersion3_3_2(t *testing.T) {
	var bcryptHash1, bcryptHash2, result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "bcrypt_hash", &bcryptHash1),
					testExtractResourceAttr("random_password.test", "result", &result1),
					testBcryptHashValid(&bcryptHash1, &result1),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "bcrypt_hash", &bcryptHash2),
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&bcryptHash1, &bcryptHash2),
					testBcryptHashValid(&bcryptHash2, &result2),
				),
			},
		},
	})
}

// TestAccResourcePassword_BcryptHash_FromVersion3_4_2 verifies behaviour when
// upgrading state from schema V2 to V3 with an expected bcrypt_hash update.
func TestAccResourcePassword_BcryptHash_FromVersion3_4_2(t *testing.T) {
	var bcryptHash1, bcryptHash2, result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion342(),
				Config: `resource "random_password" "test" {
							length = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "bcrypt_hash", &bcryptHash1),
					testExtractResourceAttr("random_password.test", "result", &result1),
					testBcryptHashInvalid(&bcryptHash1, &result1),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "bcrypt_hash", &bcryptHash2),
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesDiffer(&bcryptHash1, &bcryptHash2),
					testBcryptHashValid(&bcryptHash2, &result2),
				),
			},
		},
	})
}

func TestAccResourcePassword_Override(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
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

// TestAccResourcePassword_StateUpgradeV0toV2 covers the state upgrades from V0 to V2.
// This includes the the addition of `numeric` and `bcrypt_hash` attributes.
// v3.1.3 is used as this is last version before `bcrypt_hash` attributed was added.
func TestAccResourcePassword_StateUpgradeV0toV2(t *testing.T) {
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
			name: "number is absent before number and numeric are absent during",
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
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric")},
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
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
			},
		},
		{
			name: "number is true before number and numeric are absent during",
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
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
						ProtoV5ProviderFactories: protoV5ProviderFactories(),
						Config:                   c.configDuringUpgrade,
						Check:                    resource.ComposeTestCheckFunc(c.afterStateUpgrade...),
					},
				},
			})
		})
	}
}

// TestAccResourcePassword_StateUpgrade_V1toV2 covers the state upgrades from V1 to V2.
// This includes the addition of `numeric` attribute.
// v3.2.0 was used as this is the last version before `number` was deprecated and `numeric` attribute
// was added.
func TestAccResourcePassword_StateUpgradeV1toV2(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		configBeforeUpgrade string
		configDuringUpgrade string
		beforeStateUpgrade  []resource.TestCheckFunc
		afterStateUpgrade   []resource.TestCheckFunc
	}{
		{
			name: "number is absent before number and numeric are absent during",
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
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
			},
		},
		{
			name: "number is true before number and numeric are absent during",
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
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
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
						ProtoV5ProviderFactories: protoV5ProviderFactories(),
						Config:                   c.configDuringUpgrade,
						Check:                    resource.ComposeTestCheckFunc(c.afterStateUpgrade...),
					},
				},
			})
		})
	}
}

func TestAccResourcePassword_Min(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
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

// TestAccResourcePassword_UpgradeFromVersion2_2_1 verifies behaviour when upgrading state from schema V0 to V3.
func TestAccResourcePassword_UpgradeFromVersion2_2_1(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
				t.Skip("This test requires darwin/amd64 to download the old provider version. Setting TF_ACC_TERRAFORM_PATH to darwin/amd64 compatible Terraform binary can be used as a workaround.")
			}
		},
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion221(),
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
					resource.TestCheckResourceAttr("random_password.min", "special", "true"),
					resource.TestCheckResourceAttr("random_password.min", "upper", "true"),
					resource.TestCheckResourceAttr("random_password.min", "lower", "true"),
					resource.TestCheckResourceAttr("random_password.min", "number", "true"),
					resource.TestCheckResourceAttr("random_password.min", "min_special", "1"),
					resource.TestCheckResourceAttr("random_password.min", "min_upper", "3"),
					resource.TestCheckResourceAttr("random_password.min", "min_lower", "2"),
					resource.TestCheckResourceAttr("random_password.min", "min_numeric", "4"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				PlanOnly: true,
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
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
					resource.TestCheckResourceAttr("random_password.min", "special", "true"),
					resource.TestCheckResourceAttr("random_password.min", "upper", "true"),
					resource.TestCheckResourceAttr("random_password.min", "lower", "true"),
					resource.TestCheckResourceAttr("random_password.min", "number", "true"),
					resource.TestCheckResourceAttr("random_password.min", "numeric", "true"),
					resource.TestCheckResourceAttr("random_password.min", "min_special", "1"),
					resource.TestCheckResourceAttr("random_password.min", "min_upper", "3"),
					resource.TestCheckResourceAttr("random_password.min", "min_lower", "2"),
					resource.TestCheckResourceAttr("random_password.min", "min_numeric", "4"),
					resource.TestCheckResourceAttrSet("random_password.min", "bcrypt_hash"),
				),
			},
		},
	})
}

// TestAccResourcePassword_UpgradeFromVersion3_2_0 verifies behaviour when upgrading state from schema V1 to V3.
func TestAccResourcePassword_UpgradeFromVersion3_2_0(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion320(),
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
					resource.TestCheckResourceAttr("random_password.min", "special", "true"),
					resource.TestCheckResourceAttr("random_password.min", "upper", "true"),
					resource.TestCheckResourceAttr("random_password.min", "lower", "true"),
					resource.TestCheckResourceAttr("random_password.min", "number", "true"),
					resource.TestCheckResourceAttr("random_password.min", "min_special", "1"),
					resource.TestCheckResourceAttr("random_password.min", "min_upper", "3"),
					resource.TestCheckResourceAttr("random_password.min", "min_lower", "2"),
					resource.TestCheckResourceAttr("random_password.min", "min_numeric", "4"),
					resource.TestCheckResourceAttrSet("random_password.min", "bcrypt_hash"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				PlanOnly: true,
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
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
					resource.TestCheckResourceAttr("random_password.min", "special", "true"),
					resource.TestCheckResourceAttr("random_password.min", "upper", "true"),
					resource.TestCheckResourceAttr("random_password.min", "lower", "true"),
					resource.TestCheckResourceAttr("random_password.min", "number", "true"),
					resource.TestCheckResourceAttr("random_password.min", "numeric", "true"),
					resource.TestCheckResourceAttr("random_password.min", "min_special", "1"),
					resource.TestCheckResourceAttr("random_password.min", "min_upper", "3"),
					resource.TestCheckResourceAttr("random_password.min", "min_lower", "2"),
					resource.TestCheckResourceAttr("random_password.min", "min_numeric", "4"),
					resource.TestCheckResourceAttrSet("random_password.min", "bcrypt_hash"),
				),
			},
		},
	})
}

// TestAccResourcePassword_UpgradeFromVersion3_3_2 verifies behaviour when upgrading from SDKv2 to the Framework.
func TestAccResourcePassword_UpgradeFromVersion3_3_2(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
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
					resource.TestCheckResourceAttr("random_password.min", "special", "true"),
					resource.TestCheckResourceAttr("random_password.min", "upper", "true"),
					resource.TestCheckResourceAttr("random_password.min", "lower", "true"),
					resource.TestCheckResourceAttr("random_password.min", "number", "true"),
					resource.TestCheckResourceAttr("random_password.min", "numeric", "true"),
					resource.TestCheckResourceAttr("random_password.min", "min_special", "1"),
					resource.TestCheckResourceAttr("random_password.min", "min_upper", "3"),
					resource.TestCheckResourceAttr("random_password.min", "min_lower", "2"),
					resource.TestCheckResourceAttr("random_password.min", "min_numeric", "4"),
					resource.TestCheckResourceAttrSet("random_password.min", "bcrypt_hash"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "min" {
							length = 12
							override_special = "!#@"
							min_lower = 2
							min_upper = 3
							min_special = 1
							min_numeric = 4
						}`,
				PlanOnly: true,
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
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
					resource.TestCheckResourceAttr("random_password.min", "special", "true"),
					resource.TestCheckResourceAttr("random_password.min", "upper", "true"),
					resource.TestCheckResourceAttr("random_password.min", "lower", "true"),
					resource.TestCheckResourceAttr("random_password.min", "number", "true"),
					resource.TestCheckResourceAttr("random_password.min", "numeric", "true"),
					resource.TestCheckResourceAttr("random_password.min", "min_special", "1"),
					resource.TestCheckResourceAttr("random_password.min", "min_upper", "3"),
					resource.TestCheckResourceAttr("random_password.min", "min_lower", "2"),
					resource.TestCheckResourceAttr("random_password.min", "min_numeric", "4"),
					resource.TestCheckResourceAttrSet("random_password.min", "bcrypt_hash"),
				),
			},
		},
	})
}

func TestUpgradePasswordStateV0toV3(t *testing.T) {
	raw := tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "none"),
		"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"length":           tftypes.NewValue(tftypes.Number, 16),
		"lower":            tftypes.NewValue(tftypes.Bool, true),
		"min_lower":        tftypes.NewValue(tftypes.Number, 0),
		"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
		"min_special":      tftypes.NewValue(tftypes.Number, 0),
		"min_upper":        tftypes.NewValue(tftypes.Number, 0),
		"number":           tftypes.NewValue(tftypes.Bool, true),
		"override_special": tftypes.NewValue(tftypes.String, "!#$%\u0026*()-_=+[]{}\u003c\u003e:?"),
		"result":           tftypes.NewValue(tftypes.String, "DZy_3*tnonj%Q%Yx"),
		"special":          tftypes.NewValue(tftypes.Bool, true),
		"upper":            tftypes.NewValue(tftypes.Bool, true),
	})

	req := res.UpgradeStateRequest{
		State: &tfsdk.State{
			Raw:    raw,
			Schema: passwordSchemaV0(),
		},
	}

	resp := &res.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: passwordSchemaV3(),
		},
	}

	upgradePasswordStateV0toV3(context.Background(), req, resp)

	expected := passwordModelV3{
		ID:              types.String{Value: "none"},
		Keepers:         types.Map{Null: true, ElemType: types.StringType},
		Length:          types.Int64{Value: 16},
		Special:         types.Bool{Value: true},
		Upper:           types.Bool{Value: true},
		Lower:           types.Bool{Value: true},
		Number:          types.Bool{Value: true},
		Numeric:         types.Bool{Value: true},
		MinNumeric:      types.Int64{Value: 0},
		MinUpper:        types.Int64{Value: 0},
		MinLower:        types.Int64{Value: 0},
		MinSpecial:      types.Int64{Value: 0},
		OverrideSpecial: types.String{Value: "!#$%\u0026*()-_=+[]{}\u003c\u003e:?"},
		Result:          types.String{Value: "DZy_3*tnonj%Q%Yx"},
	}

	actual := passwordModelV3{}
	diags := resp.State.Get(context.Background(), &actual)
	if diags.HasError() {
		t.Errorf("error getting state: %v", diags)
	}

	err := bcrypt.CompareHashAndPassword([]byte(actual.BcryptHash.Value), []byte(actual.Result.Value))
	if err != nil {
		t.Errorf("unexpected bcrypt comparison error: %s", err)
	}

	// Setting actual.BcryptHash to zero value to allow direct comparison of expected and actual.
	actual.BcryptHash = types.String{}

	if !cmp.Equal(expected, actual) {
		t.Errorf("expected: %+v, got: %+v", expected, actual)
	}
}

func TestUpgradePasswordStateV1toV3(t *testing.T) {
	raw := tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "none"),
		"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		"length":           tftypes.NewValue(tftypes.Number, 16),
		"lower":            tftypes.NewValue(tftypes.Bool, true),
		"min_lower":        tftypes.NewValue(tftypes.Number, 0),
		"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
		"min_special":      tftypes.NewValue(tftypes.Number, 0),
		"min_upper":        tftypes.NewValue(tftypes.Number, 0),
		"number":           tftypes.NewValue(tftypes.Bool, true),
		"override_special": tftypes.NewValue(tftypes.String, "!#$%\u0026*()-_=+[]{}\u003c\u003e:?"),
		"result":           tftypes.NewValue(tftypes.String, "DZy_3*tnonj%Q%Yx"),
		"special":          tftypes.NewValue(tftypes.Bool, true),
		"upper":            tftypes.NewValue(tftypes.Bool, true),
		"bcrypt_hash":      tftypes.NewValue(tftypes.String, "bcrypt_hash"),
	})

	req := res.UpgradeStateRequest{
		State: &tfsdk.State{
			Raw:    raw,
			Schema: passwordSchemaV1(),
		},
	}

	resp := &res.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: passwordSchemaV3(),
		},
	}

	upgradePasswordStateV1toV3(context.Background(), req, resp)

	expected := passwordModelV3{
		ID:              types.String{Value: "none"},
		Keepers:         types.Map{Null: true, ElemType: types.StringType},
		Length:          types.Int64{Value: 16},
		Special:         types.Bool{Value: true},
		Upper:           types.Bool{Value: true},
		Lower:           types.Bool{Value: true},
		Number:          types.Bool{Value: true},
		Numeric:         types.Bool{Value: true},
		MinNumeric:      types.Int64{Value: 0},
		MinUpper:        types.Int64{Value: 0},
		MinLower:        types.Int64{Value: 0},
		MinSpecial:      types.Int64{Value: 0},
		OverrideSpecial: types.String{Value: "!#$%\u0026*()-_=+[]{}\u003c\u003e:?"},
		BcryptHash:      types.String{Value: "bcrypt_hash"},
		Result:          types.String{Value: "DZy_3*tnonj%Q%Yx"},
	}

	actual := passwordModelV3{}
	diags := resp.State.Get(context.Background(), &actual)
	if diags.HasError() {
		t.Errorf("error getting state: %v", diags)
	}

	if !cmp.Equal(expected, actual) {
		t.Errorf("expected: %+v, got: %+v", expected, actual)
	}
}

func TestUpgradePasswordStateV2toV3(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  res.UpgradeStateRequest
		expected *res.UpgradeStateResponse
	}{
		"valid-hash": {
			request: res.UpgradeStateRequest{
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bcrypt_hash":      tftypes.String,
							"id":               tftypes.String,
							"keepers":          tftypes.Map{ElementType: tftypes.String},
							"length":           tftypes.Number,
							"lower":            tftypes.Bool,
							"min_lower":        tftypes.Number,
							"min_numeric":      tftypes.Number,
							"min_special":      tftypes.Number,
							"min_upper":        tftypes.Number,
							"number":           tftypes.Bool,
							"numeric":          tftypes.Bool,
							"override_special": tftypes.String,
							"result":           tftypes.String,
							"special":          tftypes.Bool,
							"upper":            tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"bcrypt_hash":      tftypes.NewValue(tftypes.String, "$2a$10$d9zhEkVg.O1jZ6fEIMRlRuu/vMa0/4UIzeK5joaTBhZJlYiIPhWWa"),
						"id":               tftypes.NewValue(tftypes.String, "none"),
						"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"length":           tftypes.NewValue(tftypes.Number, 20),
						"lower":            tftypes.NewValue(tftypes.Bool, true),
						"min_lower":        tftypes.NewValue(tftypes.Number, 0),
						"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
						"min_special":      tftypes.NewValue(tftypes.Number, 0),
						"min_upper":        tftypes.NewValue(tftypes.Number, 0),
						"number":           tftypes.NewValue(tftypes.Bool, true),
						"numeric":          tftypes.NewValue(tftypes.Bool, true),
						"override_special": tftypes.NewValue(tftypes.String, ""),
						"result":           tftypes.NewValue(tftypes.String, "n:um[a9kO&x!L=9og[EM"),
						"special":          tftypes.NewValue(tftypes.Bool, true),
						"upper":            tftypes.NewValue(tftypes.Bool, true),
					}),
					Schema: passwordSchemaV2(),
				},
			},
			expected: &res.UpgradeStateResponse{
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bcrypt_hash":      tftypes.String,
							"id":               tftypes.String,
							"keepers":          tftypes.Map{ElementType: tftypes.String},
							"length":           tftypes.Number,
							"lower":            tftypes.Bool,
							"min_lower":        tftypes.Number,
							"min_numeric":      tftypes.Number,
							"min_special":      tftypes.Number,
							"min_upper":        tftypes.Number,
							"number":           tftypes.Bool,
							"numeric":          tftypes.Bool,
							"override_special": tftypes.String,
							"result":           tftypes.String,
							"special":          tftypes.Bool,
							"upper":            tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						// The difference checking should compare this actual
						// value since it should not be updated.
						"bcrypt_hash":      tftypes.NewValue(tftypes.String, "$2a$10$d9zhEkVg.O1jZ6fEIMRlRuu/vMa0/4UIzeK5joaTBhZJlYiIPhWWa"),
						"id":               tftypes.NewValue(tftypes.String, "none"),
						"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"length":           tftypes.NewValue(tftypes.Number, 20),
						"lower":            tftypes.NewValue(tftypes.Bool, true),
						"min_lower":        tftypes.NewValue(tftypes.Number, 0),
						"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
						"min_special":      tftypes.NewValue(tftypes.Number, 0),
						"min_upper":        tftypes.NewValue(tftypes.Number, 0),
						"number":           tftypes.NewValue(tftypes.Bool, true),
						"numeric":          tftypes.NewValue(tftypes.Bool, true),
						"override_special": tftypes.NewValue(tftypes.String, ""),
						"result":           tftypes.NewValue(tftypes.String, "n:um[a9kO&x!L=9og[EM"),
						"special":          tftypes.NewValue(tftypes.Bool, true),
						"upper":            tftypes.NewValue(tftypes.Bool, true),
					}),
					Schema: passwordSchemaV3(),
				},
			},
		},
		"invalid-hash": {
			request: res.UpgradeStateRequest{
				State: &tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bcrypt_hash":      tftypes.String,
							"id":               tftypes.String,
							"keepers":          tftypes.Map{ElementType: tftypes.String},
							"length":           tftypes.Number,
							"lower":            tftypes.Bool,
							"min_lower":        tftypes.Number,
							"min_numeric":      tftypes.Number,
							"min_special":      tftypes.Number,
							"min_upper":        tftypes.Number,
							"number":           tftypes.Bool,
							"numeric":          tftypes.Bool,
							"override_special": tftypes.String,
							"result":           tftypes.String,
							"special":          tftypes.Bool,
							"upper":            tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"bcrypt_hash":      tftypes.NewValue(tftypes.String, "$2a$10$bPOZGBpGe4XIgbpVaWNya.dz/HsU1GDLjuEposH2wf.vUO5rA1wXe"),
						"id":               tftypes.NewValue(tftypes.String, "none"),
						"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"length":           tftypes.NewValue(tftypes.Number, 20),
						"lower":            tftypes.NewValue(tftypes.Bool, true),
						"min_lower":        tftypes.NewValue(tftypes.Number, 0),
						"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
						"min_special":      tftypes.NewValue(tftypes.Number, 0),
						"min_upper":        tftypes.NewValue(tftypes.Number, 0),
						"number":           tftypes.NewValue(tftypes.Bool, true),
						"numeric":          tftypes.NewValue(tftypes.Bool, true),
						"override_special": tftypes.NewValue(tftypes.String, ""),
						"result":           tftypes.NewValue(tftypes.String, "$7r>NiN4Z%uAxpU]:DuB"),
						"special":          tftypes.NewValue(tftypes.Bool, true),
						"upper":            tftypes.NewValue(tftypes.Bool, true),
					}),
					Schema: passwordSchemaV2(),
				},
			},
			expected: &res.UpgradeStateResponse{
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bcrypt_hash":      tftypes.String,
							"id":               tftypes.String,
							"keepers":          tftypes.Map{ElementType: tftypes.String},
							"length":           tftypes.Number,
							"lower":            tftypes.Bool,
							"min_lower":        tftypes.Number,
							"min_numeric":      tftypes.Number,
							"min_special":      tftypes.Number,
							"min_upper":        tftypes.Number,
							"number":           tftypes.Bool,
							"numeric":          tftypes.Bool,
							"override_special": tftypes.String,
							"result":           tftypes.String,
							"special":          tftypes.Bool,
							"upper":            tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						// bcrypt_hash is randomly generated, so the difference checking
						// will ignore this value.
						"bcrypt_hash":      tftypes.NewValue(tftypes.String, nil),
						"id":               tftypes.NewValue(tftypes.String, "none"),
						"keepers":          tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						"length":           tftypes.NewValue(tftypes.Number, 20),
						"lower":            tftypes.NewValue(tftypes.Bool, true),
						"min_lower":        tftypes.NewValue(tftypes.Number, 0),
						"min_numeric":      tftypes.NewValue(tftypes.Number, 0),
						"min_special":      tftypes.NewValue(tftypes.Number, 0),
						"min_upper":        tftypes.NewValue(tftypes.Number, 0),
						"number":           tftypes.NewValue(tftypes.Bool, true),
						"numeric":          tftypes.NewValue(tftypes.Bool, true),
						"override_special": tftypes.NewValue(tftypes.String, ""),
						"result":           tftypes.NewValue(tftypes.String, "$7r>NiN4Z%uAxpU]:DuB"),
						"special":          tftypes.NewValue(tftypes.Bool, true),
						"upper":            tftypes.NewValue(tftypes.Bool, true),
					}),
					Schema: passwordSchemaV3(),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := res.UpgradeStateResponse{
				State: tfsdk.State{
					Schema: testCase.expected.State.Schema,
				},
			}

			upgradePasswordStateV2toV3(context.Background(), testCase.request, &got)

			// Since bcrypt_hash is generated, this test is very involved to
			// ensure the test case is set up properly and the generated
			// value is removed to prevent false positive differences.
			var err error
			var requestBcryptHash, requestResult, expectedBcryptHash, gotBcryptHash, gotResult string

			bcryptHashPath := tftypes.NewAttributePath().WithAttributeName("bcrypt_hash")
			resultPath := tftypes.NewAttributePath().WithAttributeName("result")

			requestBcryptHashValue, err := testTftypesValueAtPath(testCase.request.State.Raw, bcryptHashPath)

			if err != nil {
				t.Fatalf("unexpected error getting request bcrypt_hash value: %s", err)
			}

			if err := requestBcryptHashValue.As(&requestBcryptHash); err != nil {
				t.Fatalf("unexpected error converting request bcrypt_hash to string: %s", err)
			}

			requestResultValue, err := testTftypesValueAtPath(testCase.request.State.Raw, resultPath)

			if err != nil {
				t.Fatalf("unexpected error getting request result value: %s", err)
			}

			if err := requestResultValue.As(&requestResult); err != nil {
				t.Fatalf("unexpected error converting request result to string: %s", err)
			}

			expectedBcryptHashValue, err := testTftypesValueAtPath(testCase.expected.State.Raw, bcryptHashPath)

			if err != nil {
				t.Fatalf("unexpected error getting expected bcrypt_hash value: %s", err)
			}

			if err := expectedBcryptHashValue.As(&expectedBcryptHash); err != nil {
				t.Fatalf("unexpected error converting expected bcrypt_hash to string: %s", err)
			}

			gotBcryptHashValue, err := testTftypesValueAtPath(got.State.Raw, bcryptHashPath)

			if err != nil {
				t.Fatalf("unexpected error getting got bcrypt_hash value: %s", err)
			}

			if err := gotBcryptHashValue.As(&gotBcryptHash); err != nil {
				t.Fatalf("unexpected error converting got bcrypt_hash to string: %s", err)
			}

			gotResultValue, err := testTftypesValueAtPath(got.State.Raw, resultPath)

			if err != nil {
				t.Fatalf("unexpected error getting got result value: %s", err)
			}

			if err := gotResultValue.As(&gotResult); err != nil {
				t.Fatalf("unexpected error converting got result to string: %s", err)
			}

			err = bcrypt.CompareHashAndPassword([]byte(requestBcryptHash), []byte(requestResult))

			// If the request bcrypt_hash was valid, it should be in expected
			// and got. Otherwise, it should be regenerated which will be a
			// random value which must be stripped to prevent false positives.
			if err == nil {
				// Ensure the test case is valid.
				if !requestBcryptHashValue.Equal(expectedBcryptHashValue) {
					t.Fatal("expected request bcrypt_hash in expected")
				}

				// Ensure the request bcrypt_hash was not modified.
				if !requestBcryptHashValue.Equal(gotBcryptHashValue) {
					t.Fatal("expected request bcrypt_hash in got")
				}
			} else {
				// If we got a different error than mismatched hash, then the
				// test case might not be valid.
				if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
					t.Fatalf("unexpected request bcrypt_hash error: %s", err)
				}

				// Ensure the test case has null values on both sides as a
				// regenerated bcrypt_hash cannot be equality compared.
				if !expectedBcryptHashValue.IsNull() {
					t.Fatal("expected null bcrypt_hash in expected")
				}

				// Prevent differences from the got bcrypt_path being randomly
				// generated.
				got.State.Raw, err = tftypes.Transform(
					got.State.Raw,
					func(path *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
						// Purposefully set bcrypt_hash value to nil.
						if path.Equal(bcryptHashPath) {
							return tftypes.NewValue(tftypes.String, nil), nil
						}

						return value, nil
					},
				)

				if err != nil {
					t.Fatalf("unexpected error transforming got: %s", err)
				}
			}

			// The got bcrypt_hash should always be valid.
			if err := bcrypt.CompareHashAndPassword([]byte(gotBcryptHash), []byte(gotResult)); err != nil {
				t.Errorf("unexpected error comparing got bcrypt_hash and result: %s", err)
			}

			// Ensure all state values are checked.
			if diff := cmp.Diff(*testCase.expected, got); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAccResourcePassword_NumberNumericErrors(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "number_numeric_differ" {
  							length = 1
							number = false
  							numeric = true
						}`,
				ExpectError: regexp.MustCompile(`.*Number and numeric are both configured with different values`),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_EmptyMap(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_EmptyMapToNullValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_NullMap(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_NullMapToNullValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_NullValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_NullValues(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "2"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "2"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_Value(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Keep_Values(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "2"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "2"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_EmptyMapToValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesDiffer(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_NullMapToValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesDiffer(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_NullValueToValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesDiffer(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_ValueToEmptyMap(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesDiffer(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_ValueToNullMap(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesDiffer(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_ValueToNullValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesDiffer(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_Replace_ValueToNewValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "456"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesDiffer(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_FrameworkMigration_NullMapToNullValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_FrameworkMigration_NullMapToValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key" = "123"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesDiffer(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_FrameworkMigration_NullMapToMultipleNullValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "2"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_FrameworkMigration_NullMapToMultipleValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = null
						"key2" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesDiffer(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "2"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_FrameworkMigration_NullMapValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesEqual(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "2"),
				),
			},
		},
	})
}

func TestAccResourcePassword_Keepers_FrameworkMigration_NullMapValueToValue(t *testing.T) {
	var result1, result2 string

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion332(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = null
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result1),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config: `resource "random_password" "test" {
					length = 12
					keepers = {
						"key1" = "123"
						"key2" = "456"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testExtractResourceAttr("random_password.test", "result", &result2),
					testCheckAttributeValuesDiffer(&result1, &result2),
					resource.TestCheckResourceAttr("random_password.test", "keepers.%", "2"),
				),
			},
		},
	})
}

func testBcryptHashInvalid(hash *string, password *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if hash == nil || *hash == "" {
			return fmt.Errorf("expected hash value")
		}

		if password == nil || *password == "" {
			return fmt.Errorf("expected password value")
		}

		err := bcrypt.CompareHashAndPassword([]byte(*hash), []byte(*password))

		if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return fmt.Errorf("unexpected error: %s", err)
		}

		return nil
	}
}

func testBcryptHashValid(hash *string, password *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if hash == nil || *hash == "" {
			return fmt.Errorf("expected hash value")
		}

		if password == nil || *password == "" {
			return fmt.Errorf("expected password value")
		}

		return bcrypt.CompareHashAndPassword([]byte(*hash), []byte(*password))
	}
}
