package provider_fm

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"
)

func TestResourceStringMigrateState(t *testing.T) {
	//cases := map[string]struct {
	//	StateVersion       int
	//	ID                 string
	//	InputAttributes    map[string]string
	//	ExpectedAttributes map[string]string
	//	Meta               interface{}
	//}{
	//	"v0_1_simple": {
	//		StateVersion: 0,
	//		ID:           "some_id",
	//		InputAttributes: map[string]string{
	//			"result": "foo",
	//			"id":     "foo",
	//			"length": "3",
	//		},
	//		ExpectedAttributes: map[string]string{
	//			"result":      "foo",
	//			"id":          "foo",
	//			"length":      "3",
	//			"min_numeric": "0",
	//			"min_special": "0",
	//			"min_lower":   "0",
	//			"min_upper":   "0",
	//		},
	//	},
	//	"v0_1_special": {
	//		StateVersion: 0,
	//		ID:           "some_id",
	//		InputAttributes: map[string]string{
	//			"result":           "foo",
	//			"id":               "foo",
	//			"special":          "false",
	//			"length":           "3",
	//			"override_special": "!@",
	//		},
	//		ExpectedAttributes: map[string]string{
	//			"result":           "foo",
	//			"id":               "foo",
	//			"special":          "false",
	//			"length":           "3",
	//			"override_special": "!@",
	//			"min_numeric":      "0",
	//			"min_special":      "0",
	//			"min_lower":        "0",
	//			"min_upper":        "0",
	//		},
	//	},
	//}

	tval := tftypes.Value{}

	tval.Type()

	//for tn, tc := range cases {
	req := tfsdk.UpgradeResourceStateRequest{
		State: &tfsdk.State{
			Raw: tftypes.Value{},
			Schema: tfsdk.Schema{
				Version: 0,
			},
		},
	}

	resp := tfsdk.UpgradeResourceStateResponse{}

	migrateStringStateV0toV1(context.Background(), req, &resp)

	respState := StringModel{}
	resp.State.Get(context.Background(), &respState)

	//if err != nil {
	//	t.Fatalf("bad: %s, err: %#v", tn, err)
	//}
	//
	//for k, v := range tc.ExpectedAttributes {
	//	actual := is.Attributes[k]
	//	if actual != v {
	//		t.Fatalf("Bad Random StringModel Migration for %q: %q\n\n expected: %q", k, actual, v)
	//	}
	//}
	//}
}
