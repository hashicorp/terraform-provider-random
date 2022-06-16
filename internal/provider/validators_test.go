package provider

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestIsAtLeastSumOfValidator_Validate(t *testing.T) {
	t.Parallel()

	req := tfsdk.ValidateAttributeRequest{
		AttributePath:   tftypes.NewAttributePath().WithAttributeName("length"),
		AttributeConfig: types.Int64{Value: 16},
		Config: tfsdk.Config{
			Schema: passwordSchemaV1(),
		},
	}

	cases := []struct {
		name                   string
		reqAttribConfig        attr.Value
		reqConfigRaw           tftypes.Value
		attributesToSum        []*tftypes.AttributePath
		expectDiag             bool
		expectedValidatorDiags diag.Diagnostics
	}{
		{
			name:            "attribute wrong type",
			reqAttribConfig: types.String{Value: "16"},
			expectDiag:      true,
		},
		{
			"attribute less than sum of attribute",
			types.Int64{Value: 16},
			tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"min_upper": tftypes.NewValue(tftypes.Number, 17),
			}),
			[]*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("min_upper"),
			},
			true,
			diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("length"),
					`Attribute "length" is less than summed attributes.`,
					`Attribute "length" (16) cannot be less than min_upper (17).`,
				),
			},
		},
		{
			"attribute less than sum of attributes",
			types.Int64{Value: 16},
			tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"min_upper": tftypes.NewValue(tftypes.Number, 10),
				"min_lower": tftypes.NewValue(tftypes.Number, 12),
			}),
			[]*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("min_upper"),
				tftypes.NewAttributePath().WithAttributeName("min_lower"),
			},
			true,
			diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("length"),
					`Attribute "length" is less than summed attributes.`,
					`Attribute "length" (16) cannot be less than min_upper + min_lower (22).`,
				),
			},
		},
		{
			"a summed attribute is of invalid type",
			types.Int64{Value: 16},
			tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"min_upper": tftypes.NewValue(tftypes.String, "17"),
			}),
			[]*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("min_upper"),
			},
			true,
			diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("min_upper"),
					`Int64 Type Validation Error`,
					`An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:

Expected Number value, received tftypes.Value with value: tftypes.String<"17">`,
				),
			},
		},
		{
			name:            "attribute equal to sum of attributes",
			reqAttribConfig: types.Int64{Value: 16},
			reqConfigRaw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"min_upper": tftypes.NewValue(tftypes.Number, 8),
				"min_lower": tftypes.NewValue(tftypes.Number, 8),
			}),
			attributesToSum: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("min_upper"),
				tftypes.NewAttributePath().WithAttributeName("min_lower"),
			},
			expectDiag: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req.AttributeConfig = c.reqAttribConfig
			req.Config.Raw = c.reqConfigRaw
			resp := tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{},
			}

			validator := NewIntIsAtLeastSumOfValidator(c.attributesToSum...)
			validator.Validate(context.Background(), req, &resp)

			if c.expectDiag {
				if len(resp.Diagnostics) != 1 {
					t.Errorf("expecting resp diags len: 1, actual resp diags len: %d", len(resp.Diagnostics))
				}
			}

			// Only test the contents of diags that are explicitly under the control of the validator.
			if c.expectedValidatorDiags != nil {
				if !cmp.Equal(c.expectedValidatorDiags, resp.Diagnostics) {
					t.Errorf("expecting resp diags: %s, actual resp diags: %s", c.expectedValidatorDiags, resp.Diagnostics)
				}
			}
		})
	}
}
