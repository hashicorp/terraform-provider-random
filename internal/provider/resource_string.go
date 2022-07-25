package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-random/internal/diagnostics"
	"github.com/terraform-providers/terraform-provider-random/internal/planmodifiers"
	"github.com/terraform-providers/terraform-provider-random/internal/random"
)

var _ tfsdk.ResourceType = (*stringResourceType)(nil)

type stringResourceType struct{}

func (r stringResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
					int64validator.AtLeastSumOf(
						path.MatchRoot("min_upper"),
						path.MatchRoot("min_lower"),
						path.MatchRoot("min_numeric"),
						path.MatchRoot("min_special"),
					),
				},
			},

			"special": {
				Description: "Include special characters in the result. These are `!@#$%&*()-_=+[]{}<>:?`. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Bool{Value: true}),
					planmodifiers.RequiresReplace(),
				},
			},

			"upper": {
				Description: "Include uppercase alphabet characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Bool{Value: true}),
					planmodifiers.RequiresReplace(),
				},
			},

			"lower": {
				Description: "Include lowercase alphabet characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Bool{Value: true}),
					planmodifiers.RequiresReplace(),
				},
			},

			"numeric": {
				Description: "Include numeric characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Bool{Value: true}),
					planmodifiers.RequiresReplace(),
				},
			},

			"min_numeric": {
				Description: "Minimum number of numeric characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Int64{Value: 0}),
					planmodifiers.RequiresReplace(),
				},
			},

			"min_upper": {
				Description: "Minimum number of uppercase alphabet characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Int64{Value: 0}),
					planmodifiers.RequiresReplace(),
				},
			},

			"min_lower": {
				Description: "Minimum number of lowercase alphabet characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Int64{Value: 0}),
					planmodifiers.RequiresReplace(),
				},
			},

			"min_special": {
				Description: "Minimum number of special characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Int64{Value: 0}),
					planmodifiers.RequiresReplace(),
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
	}, nil
}

func (r stringResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return &stringResource{}, nil
}

var (
	_ tfsdk.Resource                 = (*stringResource)(nil)
	_ tfsdk.ResourceWithImportState  = (*stringResource)(nil)
	_ tfsdk.ResourceWithUpgradeState = (*stringResource)(nil)
)

type stringResource struct{}

func (r *stringResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan stringModelV2

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := random.StringParams{
		Length:          plan.Length.Value,
		Upper:           plan.Upper.Value,
		MinUpper:        plan.MinUpper.Value,
		Lower:           plan.Lower.Value,
		MinLower:        plan.MinLower.Value,
		Numeric:         plan.Numeric.Value,
		MinNumeric:      plan.MinNumeric.Value,
		Special:         plan.Special.Value,
		MinSpecial:      plan.MinSpecial.Value,
		OverrideSpecial: plan.OverrideSpecial.Value,
	}

	result, err := random.CreateString(params)
	if err != nil {
		resp.Diagnostics.Append(diagnostics.RandomReadError(err.Error())...)
		return
	}

	state := stringModelV2{
		ID:              types.String{Value: string(result)},
		Keepers:         plan.Keepers,
		Length:          types.Int64{Value: plan.Length.Value},
		Special:         types.Bool{Value: plan.Special.Value},
		Upper:           types.Bool{Value: plan.Upper.Value},
		Lower:           types.Bool{Value: plan.Lower.Value},
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

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *stringResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}

// Update is intentionally left blank as all required and optional attributes force replacement of the resource
// through the RequiresReplace AttributePlanModifier.
func (r *stringResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *stringResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

func (r *stringResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	id := req.ID

	state := stringModelV2{
		ID:         types.String{Value: id},
		Result:     types.String{Value: id},
		Length:     types.Int64{Value: int64(len(id))},
		Special:    types.Bool{Value: true},
		Upper:      types.Bool{Value: true},
		Lower:      types.Bool{Value: true},
		Numeric:    types.Bool{Value: true},
		MinSpecial: types.Int64{Value: 0},
		MinUpper:   types.Int64{Value: 0},
		MinLower:   types.Int64{Value: 0},
		MinNumeric: types.Int64{Value: 0},
	}

	state.Keepers.ElemType = types.StringType

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *stringResource) UpgradeState(context.Context) map[int64]tfsdk.ResourceStateUpgrader {
	schemaV1 := tfsdk.Schema{
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
					int64validator.AtLeastSumOf(
						path.MatchRoot("min_upper"),
						path.MatchRoot("min_lower"),
						path.MatchRoot("min_numeric"),
						path.MatchRoot("min_special"),
					),
				},
			},

			"special": {
				Description: "Include special characters in the result. These are `!@#$%&*()-_=+[]{}<>:?`. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Bool{Value: true}),
					planmodifiers.RequiresReplace(),
				},
			},

			"upper": {
				Description: "Include uppercase alphabet characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Bool{Value: true}),
					planmodifiers.RequiresReplace(),
				},
			},

			"lower": {
				Description: "Include lowercase alphabet characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Bool{Value: true}),
					planmodifiers.RequiresReplace(),
				},
			},

			"number": {
				Description: "Include numeric characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Bool{Value: true}),
					planmodifiers.RequiresReplace(),
				},
			},

			"min_numeric": {
				Description: "Minimum number of numeric characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Int64{Value: 0}),
					planmodifiers.RequiresReplace(),
				},
			},

			"min_upper": {
				Description: "Minimum number of uppercase alphabet characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Int64{Value: 0}),
					planmodifiers.RequiresReplace(),
				},
			},

			"min_lower": {
				Description: "Minimum number of lowercase alphabet characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Int64{Value: 0}),
					planmodifiers.RequiresReplace(),
				},
			},

			"min_special": {
				Description: "Minimum number of special characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Int64{Value: 0}),
					planmodifiers.RequiresReplace(),
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

	return map[int64]tfsdk.ResourceStateUpgrader{
		1: {
			PriorSchema:   &schemaV1,
			StateUpgrader: upgradeStringStateV1toV2,
		},
	}
}

func upgradeStringStateV1toV2(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
	type modelV1 struct {
		ID              types.String `tfsdk:"id"`
		Keepers         types.Map    `tfsdk:"keepers"`
		Length          types.Int64  `tfsdk:"length"`
		Special         types.Bool   `tfsdk:"special"`
		Upper           types.Bool   `tfsdk:"upper"`
		Lower           types.Bool   `tfsdk:"lower"`
		Number          types.Bool   `tfsdk:"number"`
		MinNumeric      types.Int64  `tfsdk:"min_numeric"`
		MinUpper        types.Int64  `tfsdk:"min_upper"`
		MinLower        types.Int64  `tfsdk:"min_lower"`
		MinSpecial      types.Int64  `tfsdk:"min_special"`
		OverrideSpecial types.String `tfsdk:"override_special"`
		Result          types.String `tfsdk:"result"`
	}

	var stringDataV1 modelV1

	resp.Diagnostics.Append(req.State.Get(ctx, &stringDataV1)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stringDataV2 := stringModelV2{
		Keepers:         stringDataV1.Keepers,
		Length:          stringDataV1.Length,
		Special:         stringDataV1.Special,
		Upper:           stringDataV1.Upper,
		Lower:           stringDataV1.Lower,
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

type stringModelV2 struct {
	ID              types.String `tfsdk:"id"`
	Keepers         types.Map    `tfsdk:"keepers"`
	Length          types.Int64  `tfsdk:"length"`
	Special         types.Bool   `tfsdk:"special"`
	Upper           types.Bool   `tfsdk:"upper"`
	Lower           types.Bool   `tfsdk:"lower"`
	Numeric         types.Bool   `tfsdk:"numeric"`
	MinNumeric      types.Int64  `tfsdk:"min_numeric"`
	MinUpper        types.Int64  `tfsdk:"min_upper"`
	MinLower        types.Int64  `tfsdk:"min_lower"`
	MinSpecial      types.Int64  `tfsdk:"min_special"`
	OverrideSpecial types.String `tfsdk:"override_special"`
	Result          types.String `tfsdk:"result"`
}
