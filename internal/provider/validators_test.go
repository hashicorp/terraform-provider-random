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

func TestIntAtLeastValidator_Validate(t *testing.T) {
	t.Parallel()

	req := tfsdk.ValidateAttributeRequest{
		AttributePath: tftypes.NewAttributePath().WithAttributeName("length"),
		Config: tfsdk.Config{
			Schema: getPasswordSchemaV1(),
		},
	}

	type expectedRespDiags struct {
		expectedRespDiagAttrPath *tftypes.AttributePath
		expectedRespDiagSummary  string
		expectedRespDiagDetail   string
	}

	cases := []struct {
		name              string
		reqAttribConfig   attr.Value
		reqConfigRaw      tftypes.Value
		attributesToSum   []*tftypes.AttributePath
		expectedRespDiags []expectedRespDiags
	}{
		{
			name:            "attribute wrong type",
			reqAttribConfig: types.String{Value: "16"},
			expectedRespDiags: []expectedRespDiags{
				{
					tftypes.NewAttributePath().WithAttributeName("length"),
					`Attribute "length" is of incorrect type for validator.`,
					`Attribute "length" (types.StringType) cannot be used as types.Int64Type.`,
				},
			},
		},
		{
			name:            "attribute less than min val",
			reqAttribConfig: types.Int64{Value: 5},
			expectedRespDiags: []expectedRespDiags{
				{
					tftypes.NewAttributePath().WithAttributeName("length"),
					`Attribute "length" is less than minimum required.`,
					`Attribute "length" (5) must be at least 10.`,
				},
			},
		},
		{
			name:              "attribute equal to min val",
			reqAttribConfig:   types.Int64{Value: 10},
			expectedRespDiags: []expectedRespDiags{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req.AttributeConfig = c.reqAttribConfig
			req.Config.Raw = c.reqConfigRaw
			resp := tfsdk.ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{},
			}

			validator := NewIntAtLeastValidator(10)
			validator.Validate(context.Background(), req, &resp)

			expectedDiags := diag.Diagnostics{}

			for _, v := range c.expectedRespDiags {
				expectedDiags.AddAttributeError(v.expectedRespDiagAttrPath, v.expectedRespDiagSummary, v.expectedRespDiagDetail)
			}

			if !cmp.Equal(expectedDiags, resp.Diagnostics) {
				t.Errorf("expecting resp diags: %s, actual resp diags: %s", expectedDiags, resp.Diagnostics)
			}
		})
	}
}

func TestIsAtLeastSumOfValidator_Validate(t *testing.T) {
	t.Parallel()

	req := tfsdk.ValidateAttributeRequest{
		AttributePath:   tftypes.NewAttributePath().WithAttributeName("length"),
		AttributeConfig: types.Int64{Value: 16},
		Config: tfsdk.Config{
			Schema: getPasswordSchemaV1(),
		},
	}

	type expectedRespDiags struct {
		expectedRespDiagAttrPath *tftypes.AttributePath
		expectedRespDiagSummary  string
		expectedRespDiagDetail   string
	}

	cases := []struct {
		name              string
		reqAttribConfig   attr.Value
		reqConfigRaw      tftypes.Value
		attributesToSum   []*tftypes.AttributePath
		expectedRespDiags []expectedRespDiags
	}{
		{
			name:            "attribute wrong type",
			reqAttribConfig: types.String{Value: "16"},
			expectedRespDiags: []expectedRespDiags{
				{
					tftypes.NewAttributePath().WithAttributeName("length"),
					`Attribute "length" is of incorrect type for validator.`,
					`Attribute "length" (types.StringType) cannot be used as types.Int64Type.`,
				},
			},
		},
		{
			"attribute less than sum of attribute",
			types.Int64{Value: 16},
			tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"min_upper": tftypes.NewValue(tftypes.Number, 17),
			}),
			[]*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("min_upper")},
			[]expectedRespDiags{
				{
					tftypes.NewAttributePath().WithAttributeName("length"),
					`Attribute "length" is less than summed attributes.`,
					`Attribute "length" (16) cannot be less than min_upper (17).`,
				},
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
			[]expectedRespDiags{
				{
					tftypes.NewAttributePath().WithAttributeName("length"),
					`Attribute "length" is less than summed attributes.`,
					`Attribute "length" (16) cannot be less than min_upper + min_lower (22).`,
				},
			},
		},
		{
			"a summed attribute is of invalid type",
			types.Int64{Value: 16},
			tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"min_upper": tftypes.NewValue(tftypes.String, "17"),
			}),
			[]*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("min_upper")},
			[]expectedRespDiags{
				{
					tftypes.NewAttributePath().WithAttributeName("min_upper"),
					`Int64 Type Validation Error`,
					`An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:

Expected Number value, received tftypes.Value with value: tftypes.String<"17">`,
				},
			},
		},
		{
			"attribute equal to sum of attributes",
			types.Int64{Value: 16},
			tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"min_upper": tftypes.NewValue(tftypes.Number, 8),
				"min_lower": tftypes.NewValue(tftypes.Number, 8),
			}),
			[]*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("min_upper"),
				tftypes.NewAttributePath().WithAttributeName("min_lower")},
			[]expectedRespDiags{},
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

			expectedDiags := diag.Diagnostics{}

			for _, v := range c.expectedRespDiags {
				expectedDiags.AddAttributeError(v.expectedRespDiagAttrPath, v.expectedRespDiagSummary, v.expectedRespDiagDetail)
			}

			if !cmp.Equal(expectedDiags, resp.Diagnostics) {
				t.Errorf("expecting resp diags: %s, actual resp diags: %s", expectedDiags, resp.Diagnostics)
			}
		})
	}
}
