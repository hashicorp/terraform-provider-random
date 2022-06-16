package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type resourceStringType struct{}

func (r resourceStringType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return stringSchemaV2(), nil
}

func (r resourceStringType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceString{
		p: *(p.(*provider)),
	}, nil
}

type resourceString struct {
	p provider
}

func (r resourceString) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	createString(ctx, req, resp)
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r resourceString) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}

// Update is intentionally left blank as all required and optional attributes force replacement of the resource
// through the RequiresReplace AttributePlanModifier.
func (r resourceString) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r resourceString) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

func (r resourceString) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	importString(ctx, req, resp)
}

func (r resourceString) UpgradeState(context.Context) map[int64]tfsdk.ResourceStateUpgrader {
	schemaV1 := stringSchemaV1()

	return map[int64]tfsdk.ResourceStateUpgrader{
		1: {
			PriorSchema:   &schemaV1,
			StateUpgrader: upgradeStringStateV1toV2,
		},
	}
}

func stringSchemaV2() tfsdk.Schema {
	return tfsdk.Schema{
		Version: 2,
		Description: "The resource `random_string` generates a random permutation of alphanumeric " +
			"characters and optionally special characters.\n" +
			"\n" +
			"This resource *does* use a cryptographic random number generator.\n" +
			"\n" +
			"Historically this resource's intended usage has been ambiguous as the original example used " +
			"it in a password. For backwards compatibility it will continue to exist. For unique ids please " +
			"use [random_id](id.html), for sensitive random values please use [random_password](password.html).",
		Attributes: map[string]tfsdk.Attribute{
			"keepers": {
				Description: "Arbitrary map of values that, when changed, will trigger recreation of " +
					"resource. See [the main provider documentation](../index.html) for more information.",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},

			"length": {
				Description: "The length of the string desired. The minimum value for length is 1 and, length " +
					"must also be >= (`min_upper` + `min_lower` + `min_numeric` + `min_special`).",
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
				Validators: []tfsdk.AttributeValidator{
					int64validator.AtLeast(1),
					NewIntIsAtLeastSumOfValidator(
						tftypes.NewAttributePath().WithAttributeName("min_upper"),
						tftypes.NewAttributePath().WithAttributeName("min_lower"),
						tftypes.NewAttributePath().WithAttributeName("min_numeric"),
						tftypes.NewAttributePath().WithAttributeName("min_special"),
					),
				},
			},

			"special": {
				Description: "Include special characters in the result. These are `!@#$%&*()-_=+[]{}<>:?`. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					newDefaultValueAttributePlanModifier(types.Bool{Value: true}),
					RequiresReplace(),
				},
			},

			"upper": {
				Description: "Include uppercase alphabet characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					newDefaultValueAttributePlanModifier(types.Bool{Value: true}),
					RequiresReplace(),
				},
			},

			"lower": {
				Description: "Include lowercase alphabet characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					newDefaultValueAttributePlanModifier(types.Bool{Value: true}),
					RequiresReplace(),
				},
			},

			"number": {
				Description: "Include numeric characters in the result. Default value is `true`. " +
					"**NOTE**: This is deprecated, use `numeric` instead.",
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					newNumberNumericAttributePlanModifier(),
					RequiresReplace(),
				},
				DeprecationMessage: "**NOTE**: This is deprecated, use `numeric` instead.",
			},

			"numeric": {
				Description: "Include numeric characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					newNumberNumericAttributePlanModifier(),
					RequiresReplace(),
				},
			},

			"min_numeric": {
				Description: "Minimum number of numeric characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					newDefaultValueAttributePlanModifier(types.Int64{Value: 0}),
					RequiresReplace(),
				},
			},

			"min_upper": {
				Description: "Minimum number of uppercase alphabet characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					newDefaultValueAttributePlanModifier(types.Int64{Value: 0}),
					RequiresReplace(),
				},
			},

			"min_lower": {
				Description: "Minimum number of lowercase alphabet characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					newDefaultValueAttributePlanModifier(types.Int64{Value: 0}),
					RequiresReplace(),
				},
			},

			"min_special": {
				Description: "Minimum number of special characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					newDefaultValueAttributePlanModifier(types.Int64{Value: 0}),
					RequiresReplace(),
				},
			},

			"override_special": {
				Description: "Supply your own list of special characters to use for string generation.  This " +
					"overrides the default character list in the special argument.  The `special` argument must " +
					"still be set to true for any overwritten characters to be used in generation.",
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},

			"result": {
				Description: "The generated random string.",
				Type:        types.StringType,
				Computed:    true,
			},

			"id": {
				Description: "The generated random string.",
				Computed:    true,
				Type:        types.StringType,
			},
		},
	}
}

func stringSchemaV1() tfsdk.Schema {
	return tfsdk.Schema{
		Version: 1,
		Description: "The resource `random_string` generates a random permutation of alphanumeric " +
			"characters and optionally special characters.\n" +
			"\n" +
			"This resource *does* use a cryptographic random number generator.\n" +
			"\n" +
			"Historically this resource's intended usage has been ambiguous as the original example used " +
			"it in a password. For backwards compatibility it will continue to exist. For unique ids please " +
			"use [random_id](id.html), for sensitive random values please use [random_password](password.html).",
		Attributes: map[string]tfsdk.Attribute{
			"keepers": {
				Description: "Arbitrary map of values that, when changed, will trigger recreation of " +
					"resource. See [the main provider documentation](../index.html) for more information.",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},

			"length": {
				Description: "The length of the string desired. The minimum value for length is 1 and, length " +
					"must also be >= (`min_upper` + `min_lower` + `min_numeric` + `min_special`).",
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
				Validators: []tfsdk.AttributeValidator{
					int64validator.AtLeast(1),
					NewIntIsAtLeastSumOfValidator(
						tftypes.NewAttributePath().WithAttributeName("min_upper"),
						tftypes.NewAttributePath().WithAttributeName("min_lower"),
						tftypes.NewAttributePath().WithAttributeName("min_numeric"),
						tftypes.NewAttributePath().WithAttributeName("min_special"),
					),
				},
			},

			"special": {
				Description: "Include special characters in the result. These are `!@#$%&*()-_=+[]{}<>:?`. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					newDefaultValueAttributePlanModifier(types.Bool{Value: true}),
				},
			},

			"upper": {
				Description: "Include uppercase alphabet characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					newDefaultValueAttributePlanModifier(types.Bool{Value: true}),
				},
			},

			"lower": {
				Description: "Include lowercase alphabet characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					newDefaultValueAttributePlanModifier(types.Bool{Value: true}),
				},
			},

			"number": {
				Description: "Include numeric characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					newDefaultValueAttributePlanModifier(types.Bool{Value: true}),
				},
			},

			"min_numeric": {
				Description: "Minimum number of numeric characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					newDefaultValueAttributePlanModifier(types.Int64{Value: 0}),
				},
			},

			"min_upper": {
				Description: "Minimum number of uppercase alphabet characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					newDefaultValueAttributePlanModifier(types.Int64{Value: 0}),
				},
			},

			"min_lower": {
				Description: "Minimum number of lowercase alphabet characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					newDefaultValueAttributePlanModifier(types.Int64{Value: 0}),
				},
			},

			"min_special": {
				Description: "Minimum number of special characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					newDefaultValueAttributePlanModifier(types.Int64{Value: 0}),
				},
			},

			"override_special": {
				Description: "Supply your own list of special characters to use for string generation.  This " +
					"overrides the default character list in the special argument.  The `special` argument must " +
					"still be set to true for any overwritten characters to be used in generation.",
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},

			"result": {
				Description: "The generated random string.",
				Type:        types.StringType,
				Computed:    true,
			},

			"id": {
				Description: "The generated random string.",
				Computed:    true,
				Type:        types.StringType,
			},
		},
	}
}

func createString(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan StringModelV2

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := randomStringParams{
		length:          plan.Length.Value,
		upper:           plan.Upper.Value,
		minUpper:        plan.MinUpper.Value,
		lower:           plan.Lower.Value,
		minLower:        plan.MinLower.Value,
		numeric:         plan.Numeric.Value,
		minNumeric:      plan.MinNumeric.Value,
		special:         plan.Special.Value,
		minSpecial:      plan.MinSpecial.Value,
		overrideSpecial: plan.OverrideSpecial.Value,
	}

	result, err := createRandomString(params)
	if err != nil {
		resp.Diagnostics.Append(randomReadError(err.Error())...)
		return
	}

	state := StringModelV2{
		ID:              types.String{Value: string(result)},
		Keepers:         plan.Keepers,
		Length:          types.Int64{Value: plan.Length.Value},
		Special:         types.Bool{Value: plan.Special.Value},
		Upper:           types.Bool{Value: plan.Upper.Value},
		Lower:           types.Bool{Value: plan.Lower.Value},
		Number:          types.Bool{Value: plan.Number.Value},
		Numeric:         types.Bool{Value: plan.Numeric.Value},
		MinNumeric:      types.Int64{Value: plan.MinNumeric.Value},
		MinUpper:        types.Int64{Value: plan.MinUpper.Value},
		MinLower:        types.Int64{Value: plan.MinLower.Value},
		MinSpecial:      types.Int64{Value: plan.MinSpecial.Value},
		OverrideSpecial: types.String{Value: plan.OverrideSpecial.Value},
		Result:          types.String{Value: string(result)},
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func importString(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	id := req.ID

	state := StringModelV2{
		ID:     types.String{Value: id},
		Result: types.String{Value: id},
	}

	state.Keepers.ElemType = types.StringType

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func upgradeStringStateV1toV2(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
	var stringDataV1 StringModelV1

	resp.Diagnostics.Append(req.State.Get(ctx, &stringDataV1)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stringDataV2 := StringModelV2{
		Keepers:         stringDataV1.Keepers,
		Length:          stringDataV1.Length,
		Special:         stringDataV1.Special,
		Upper:           stringDataV1.Upper,
		Lower:           stringDataV1.Lower,
		Number:          stringDataV1.Number,
		Numeric:         stringDataV1.Number,
		MinNumeric:      stringDataV1.MinNumeric,
		MinLower:        stringDataV1.MinLower,
		MinSpecial:      stringDataV1.MinSpecial,
		OverrideSpecial: stringDataV1.OverrideSpecial,
		Result:          stringDataV1.Result,
		ID:              stringDataV1.ID,
	}

	diags := resp.State.Set(ctx, stringDataV2)
	resp.Diagnostics.Append(diags...)
}
