package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
				ImportStateVerifyIgnore: []string{"bcrypt_hash", "length", "lower", "number", "special", "upper", "min_lower", "min_numeric", "min_special", "min_upper", "override_special"},
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
							number = false
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

func TestMigratePasswordStateV0toV1(t *testing.T) {
	rawStateJSON := `{
            "id": "none",
            "keepers": null,
            "length": 16,
            "lower": true,
            "min_lower": 0,
            "min_numeric": 0,
            "min_special": 0,
            "min_upper": 0,
            "number": true,
            "override_special": "!#$%\u0026*()-_=+[]{}\u003c\u003e:?",
            "result": "DZy_3*tnonj%Q%Yx",
            "special": true,
            "upper": true
          }`

	req := tfsdk.UpgradeResourceStateRequest{
		RawState: &tfprotov6.RawState{
			JSON: []byte(rawStateJSON),
		},
	}

	resp := &tfsdk.UpgradeResourceStateResponse{
		State: tfsdk.State{
			Schema: getPasswordSchemaV1(),
		},
	}

	migratePasswordStateV0toV1(context.Background(), req, resp)

	val, err := resp.DynamicValue.Unmarshal(passwordDataTftypesV1())
	if err != nil {
		t.Error(err)
	}

	data := map[string]tftypes.Value{}

	err = val.As(&data)
	if err != nil {
		t.Error(err)
	}

	expectedValues := map[string]tftypes.Value{
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
	}

	for k, v := range expectedValues {
		ok := data[k].Equal(v)
		if !ok {
			t.Errorf("expected: %v, got: %v", v, data[k])
		}
	}

	var bcryptHash string
	const bcryptHashExpectedLen = 60

	err = data["bcrypt_hash"].As(&bcryptHash)
	if err != nil {
		t.Error(err)
	}

	if len(bcryptHash) != bcryptHashExpectedLen {
		t.Errorf("expected len: %v, got len: %v", bcryptHashExpectedLen, len(bcryptHash))
	}
}
