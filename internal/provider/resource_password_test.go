package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"golang.org/x/crypto/bcrypt"
)

func TestAccResourcePasswordBasic(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_password" "basic" {
							length = 12
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceStringCheck("random_password.basic", &customLens{
						customLen: 12,
					}),
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
				ImportStateVerifyIgnore: []string{"bcrypt_hash", "length", "lower", "number", "numeric", "special", "upper", "min_lower", "min_numeric", "min_special", "min_upper", "override_special"},
			},
		},
	})
}

func TestAccResourcePasswordOverride(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
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
					testAccResourceStringCheck("random_password.override", &customLens{
						customLen: 4,
					}),
					patternMatch("random_password.override", "!!!!"),
				),
			},
		},
	})
}

// TestAccResourcePassword_StateUpgraders covers the state upgrades from v0 to V2 and V1 to V2.
// This includes the addition of bcrypt_hash and numeric attributes.
func TestAccResourcePassword_StateUpgraders(t *testing.T) {
	t.Parallel()

	v1Cases := []struct {
		name                string
		configBeforeUpgrade string
		configDuringUpgrade string
		beforeStateUpgrade  []resource.TestCheckFunc
		afterStateUpgrade   []resource.TestCheckFunc
	}{
		{
			name: "%s number is absent",
			configBeforeUpgrade: `resource "random_password" "default" {
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
			name: "%s number is absent then true",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
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
			name: "%s number is absent then false",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
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
			name: "%s number is true",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
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
			name: "%s number is true then absent",
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
			name: "%s number is true then false",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			beforeStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "true"),
				resource.TestCheckNoResourceAttr("random_password.default", "numeric"),
			},
			afterStateUpgrade: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("random_password.default", "number", "false"),
				resource.TestCheckResourceAttr("random_password.default", "numeric", "false"),
				//resource.TestCheckResourceAttrPair("random_password.default", "number", "random_password.default", "numeric"),
			},
		},
		{
			name: "%s number is false",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
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
			name: "%s number is false then absent",
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
			name: "%s number is false then true",
			configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
						number = false
					}`,
			configDuringUpgrade: `resource "random_password" "default" {
						length = 12
						number = true
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

	v0Cases := v1Cases
	v0Cases = append(v0Cases, struct {
		name                string
		configBeforeUpgrade string
		configDuringUpgrade string
		beforeStateUpgrade  []resource.TestCheckFunc
		afterStateUpgrade   []resource.TestCheckFunc
	}{
		name: "%s bcrypt_hash",
		configBeforeUpgrade: `resource "random_password" "default" {
						length = 12
					}`,
		beforeStateUpgrade: []resource.TestCheckFunc{
			resource.TestCheckNoResourceAttr("random_password.default", "bcrypt_hash"),
		},
		afterStateUpgrade: []resource.TestCheckFunc{
			resource.TestCheckResourceAttrSet("random_password.default", "bcrypt_hash"),
		},
	})

	cases := map[string][]struct {
		name                string
		configBeforeUpgrade string
		configDuringUpgrade string
		beforeStateUpgrade  []resource.TestCheckFunc
		afterStateUpgrade   []resource.TestCheckFunc
	}{
		"3.1.3": v0Cases,
		"3.2.0": v1Cases,
	}

	for providerVersion, v := range cases {
		for _, c := range v {
			name := fmt.Sprintf(c.name, providerVersion)
			t.Run(name, func(t *testing.T) {
				if c.configDuringUpgrade == "" {
					c.configDuringUpgrade = c.configBeforeUpgrade
				}

				resource.UnitTest(t, resource.TestCase{
					Steps: []resource.TestStep{
						{
							ExternalProviders: map[string]resource.ExternalProvider{"random": {
								VersionConstraint: providerVersion,
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
}

func TestAccResourcePasswordMin(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
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
					testAccResourceStringCheck("random_password.min", &customLens{
						customLen: 12,
					}),
					regexMatch("random_password.min", regexp.MustCompile(`([a-z])`), 2),
					regexMatch("random_password.min", regexp.MustCompile(`([A-Z])`), 3),
					regexMatch("random_password.min", regexp.MustCompile(`([0-9])`), 4),
					regexMatch("random_password.min", regexp.MustCompile(`([!#@])`), 1),
				),
			},
		},
	})
}

func TestMigratePasswordStateV0toV2(t *testing.T) {
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

	req := tfsdk.UpgradeResourceStateRequest{
		State: &tfsdk.State{
			Raw:    raw,
			Schema: getPasswordSchemaV0(),
		},
	}

	resp := &tfsdk.UpgradeResourceStateResponse{
		State: tfsdk.State{
			Schema: getPasswordSchemaV2(),
		},
	}

	migratePasswordStateV0toV2(context.Background(), req, resp)

	expected := PasswordModelV2{
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

	actual := PasswordModelV2{}
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

func TestMigratePasswordStateV1toV2(t *testing.T) {
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

	req := tfsdk.UpgradeResourceStateRequest{
		State: &tfsdk.State{
			Raw:    raw,
			Schema: getPasswordSchemaV1(),
		},
	}

	resp := &tfsdk.UpgradeResourceStateResponse{
		State: tfsdk.State{
			Schema: getPasswordSchemaV2(),
		},
	}

	migratePasswordStateV1toV2(context.Background(), req, resp)

	expected := PasswordModelV2{
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

	actual := PasswordModelV2{}
	diags := resp.State.Get(context.Background(), &actual)
	if diags.HasError() {
		t.Errorf("error getting state: %v", diags)
	}

	if !cmp.Equal(expected, actual) {
		t.Errorf("expected: %+v, got: %+v", expected, actual)
	}
}
