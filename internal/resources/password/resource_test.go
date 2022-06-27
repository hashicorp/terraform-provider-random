package password

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"golang.org/x/crypto/bcrypt"
)

func TestUpgradePasswordStateV0toV2(t *testing.T) {
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
			Schema: schemaV0(),
		},
	}

	resp := &tfsdk.UpgradeResourceStateResponse{
		State: tfsdk.State{
			Schema: schemaV2(),
		},
	}

	upgradePasswordStateV0toV2(context.Background(), req, resp)

	expected := modelV2{
		ID:              types.String{Value: "none"},
		Keepers:         types.Map{Null: true, ElemType: types.StringType},
		Length:          types.Int64{Value: 16},
		Special:         types.Bool{Value: true},
		Upper:           types.Bool{Value: true},
		Lower:           types.Bool{Value: true},
		Numeric:         types.Bool{Value: true},
		MinNumeric:      types.Int64{Value: 0},
		MinUpper:        types.Int64{Value: 0},
		MinLower:        types.Int64{Value: 0},
		MinSpecial:      types.Int64{Value: 0},
		OverrideSpecial: types.String{Value: "!#$%\u0026*()-_=+[]{}\u003c\u003e:?"},
		Result:          types.String{Value: "DZy_3*tnonj%Q%Yx"},
	}

	actual := modelV2{}
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

func TestUpgradePasswordStateV1toV2(t *testing.T) {
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
			Schema: schemaV1(),
		},
	}

	resp := &tfsdk.UpgradeResourceStateResponse{
		State: tfsdk.State{
			Schema: schemaV2(),
		},
	}

	upgradePasswordStateV1toV2(context.Background(), req, resp)

	expected := modelV2{
		ID:              types.String{Value: "none"},
		Keepers:         types.Map{Null: true, ElemType: types.StringType},
		Length:          types.Int64{Value: 16},
		Special:         types.Bool{Value: true},
		Upper:           types.Bool{Value: true},
		Lower:           types.Bool{Value: true},
		Numeric:         types.Bool{Value: true},
		MinNumeric:      types.Int64{Value: 0},
		MinUpper:        types.Int64{Value: 0},
		MinLower:        types.Int64{Value: 0},
		MinSpecial:      types.Int64{Value: 0},
		OverrideSpecial: types.String{Value: "!#$%\u0026*()-_=+[]{}\u003c\u003e:?"},
		BcryptHash:      types.String{Value: "bcrypt_hash"},
		Result:          types.String{Value: "DZy_3*tnonj%Q%Yx"},
	}

	actual := modelV2{}
	diags := resp.State.Get(context.Background(), &actual)
	if diags.HasError() {
		t.Errorf("error getting state: %v", diags)
	}

	if !cmp.Equal(expected, actual) {
		t.Errorf("expected: %+v, got: %+v", expected, actual)
	}
}
