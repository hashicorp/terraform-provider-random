package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-random/internal/diagnostics"
	"github.com/terraform-providers/terraform-provider-random/internal/planmodifiers"
	"github.com/terraform-providers/terraform-provider-random/internal/random"
)

var _ provider.ResourceType = (*stringResourceType)(nil)

type stringResourceType struct{}

func (r stringResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return stringSchemaV3(), nil
}

func (r stringResourceType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return &stringResource{}, nil
}

var (
	_ resource.Resource                 = (*stringResource)(nil)
	_ resource.ResourceWithImportState  = (*stringResource)(nil)
	_ resource.ResourceWithUpgradeState = (*stringResource)(nil)
)

type stringResource struct{}

func (r *stringResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan stringModelV3

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

	plan.ID = types.String{Value: string(result)}
	plan.Result = types.String{Value: string(result)}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *stringResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update ensures the plan value is copied to the state to complete the update.
func (r *stringResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model stringModelV3

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *stringResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *stringResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID

	state := stringModelV3{
		ID:              types.String{Value: id},
		Result:          types.String{Value: id},
		Length:          types.Int64{Value: int64(len(id))},
		Special:         types.Bool{Value: true},
		Upper:           types.Bool{Value: true},
		Lower:           types.Bool{Value: true},
		Number:          types.Bool{Value: true},
		Numeric:         types.Bool{Value: true},
		MinSpecial:      types.Int64{Value: 0},
		MinUpper:        types.Int64{Value: 0},
		MinLower:        types.Int64{Value: 0},
		MinNumeric:      types.Int64{Value: 0},
		OverrideSpecial: types.String{Null: true},
		Keepers:         types.Map{Null: true},
	}

	state.Keepers.ElemType = types.StringType

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *stringResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	schemaV1 := stringSchemaV1()
	schemaV2 := stringSchemaV2()

	return map[int64]resource.StateUpgrader{
		1: {
			PriorSchema:   &schemaV1,
			StateUpgrader: upgradeStringStateV1toV3,
		},
		2: {
			PriorSchema:   &schemaV2,
			StateUpgrader: upgradeStringStateV2toV3,
		},
	}
}

func upgradeStringStateV1toV3(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
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

	// Setting fields that can contain null to non-null to prevent forced replacement.
	// This can occur in cases where import has been used in provider versions v3.3.1 and earlier.
	// If import has been used with v3.3.1, for instance then length, lower, number, special, upper,
	// min_lower, min_numeric, min_special and min_upper attributes will all be null in state.
	length := stringDataV1.Length

	if length.IsNull() {
		length.Null = false
		length.Value = int64(len(stringDataV1.Result.Value))
	}

	minNumeric := stringDataV1.MinNumeric

	if minNumeric.IsNull() {
		minNumeric.Null = false
	}

	minUpper := stringDataV1.MinUpper

	if minUpper.IsNull() {
		minUpper.Null = false
	}

	minLower := stringDataV1.MinLower

	if minLower.IsNull() {
		minLower.Null = false
	}

	minSpecial := stringDataV1.MinSpecial

	if minSpecial.IsNull() {
		minSpecial.Null = false
	}

	special := stringDataV1.Special

	if special.IsNull() {
		special.Null = false
		special.Value = true
	}

	upper := stringDataV1.Upper

	if upper.IsNull() {
		upper.Null = false
		upper.Value = true
	}

	lower := stringDataV1.Lower

	if lower.IsNull() {
		lower.Null = false
		lower.Value = true
	}

	number := stringDataV1.Number

	if number.IsNull() {
		number.Null = false
		number.Value = true
	}

	stringDataV3 := stringModelV3{
		Keepers:         stringDataV1.Keepers,
		Length:          length,
		Special:         special,
		Upper:           upper,
		Lower:           lower,
		Number:          number,
		Numeric:         number,
		MinNumeric:      minNumeric,
		MinUpper:        minUpper,
		MinLower:        minLower,
		MinSpecial:      minSpecial,
		OverrideSpecial: stringDataV1.OverrideSpecial,
		Result:          stringDataV1.Result,
		ID:              stringDataV1.ID,
	}

	diags := resp.State.Set(ctx, stringDataV3)
	resp.Diagnostics.Append(diags...)
}

func upgradeStringStateV2toV3(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	type modelV2 struct {
		ID              types.String `tfsdk:"id"`
		Keepers         types.Map    `tfsdk:"keepers"`
		Length          types.Int64  `tfsdk:"length"`
		Special         types.Bool   `tfsdk:"special"`
		Upper           types.Bool   `tfsdk:"upper"`
		Lower           types.Bool   `tfsdk:"lower"`
		Number          types.Bool   `tfsdk:"number"`
		Numeric         types.Bool   `tfsdk:"numeric"`
		MinNumeric      types.Int64  `tfsdk:"min_numeric"`
		MinUpper        types.Int64  `tfsdk:"min_upper"`
		MinLower        types.Int64  `tfsdk:"min_lower"`
		MinSpecial      types.Int64  `tfsdk:"min_special"`
		OverrideSpecial types.String `tfsdk:"override_special"`
		Result          types.String `tfsdk:"result"`
	}

	var stringDataV2 modelV2

	resp.Diagnostics.Append(req.State.Get(ctx, &stringDataV2)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Setting fields that can contain null to non-null to prevent forced replacement.
	// This can occur in cases where import has been used in provider versions v3.3.1 and earlier.
	// If import has been used with v3.3.1, for instance then length, lower, number, special, upper,
	// min_lower, min_numeric, min_special and min_upper attributes will all be null in state.
	length := stringDataV2.Length

	if length.IsNull() {
		length.Null = false
		length.Value = int64(len(stringDataV2.Result.Value))
	}

	minNumeric := stringDataV2.MinNumeric

	if minNumeric.IsNull() {
		minNumeric.Null = false
	}

	minUpper := stringDataV2.MinUpper

	if minUpper.IsNull() {
		minUpper.Null = false
	}

	minLower := stringDataV2.MinLower

	if minLower.IsNull() {
		minLower.Null = false
	}

	minSpecial := stringDataV2.MinSpecial

	if minSpecial.IsNull() {
		minSpecial.Null = false
	}

	special := stringDataV2.Special

	if special.IsNull() {
		special.Null = false
		special.Value = true
	}

	upper := stringDataV2.Upper

	if upper.IsNull() {
		upper.Null = false
		upper.Value = true
	}

	lower := stringDataV2.Lower

	if lower.IsNull() {
		lower.Null = false
		lower.Value = true
	}

	number := stringDataV2.Number

	if number.IsNull() {
		number.Null = false
		number.Value = true
	}

	stringDataV3 := stringModelV3{
		Keepers:         stringDataV2.Keepers,
		Length:          length,
		Special:         special,
		Upper:           upper,
		Lower:           lower,
		Number:          number,
		Numeric:         number,
		MinNumeric:      minNumeric,
		MinUpper:        minUpper,
		MinLower:        minLower,
		MinSpecial:      minSpecial,
		OverrideSpecial: stringDataV2.OverrideSpecial,
		Result:          stringDataV2.Result,
		ID:              stringDataV2.ID,
	}

	diags := resp.State.Set(ctx, stringDataV3)
	resp.Diagnostics.Append(diags...)
}

func stringSchemaV3() tfsdk.Schema {
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
					planmodifiers.RequiresReplaceIfValuesNotNull(),
				},
			},

			"length": {
				Description: "The length of the string desired. The minimum value for length is 1 and, length " +
					"must also be >= (`min_upper` + `min_lower` + `min_numeric` + `min_special`).",
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
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
				Description: "Include numeric characters in the result. Default value is `true`. " +
					"**NOTE**: This is deprecated, use `numeric` instead.",
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.NumberNumericAttributePlanModifier(),
					planmodifiers.RequiresReplace(),
				},
				DeprecationMessage: "**NOTE**: This is deprecated, use `numeric` instead.",
			},

			"numeric": {
				Description: "Include numeric characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.NumberNumericAttributePlanModifier(),
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
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplaceIf(
						planmodifiers.RequiresReplaceUnlessEmptyStringToNull(),
						"Replace on modification unless updating from empty string (\"\") to null.",
						"Replace on modification unless updating from empty string (`\"\"`) to `null`.",
					),
				},
			},

			"result": {
				Description: "The generated random string.",
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},

			"id": {
				Description: "The generated random string.",
				Computed:    true,
				Type:        types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
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
					planmodifiers.RequiresReplaceIfValuesNotNull(),
				},
			},

			"length": {
				Description: "The length of the string desired. The minimum value for length is 1 and, length " +
					"must also be >= (`min_upper` + `min_lower` + `min_numeric` + `min_special`).",
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
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
				Description: "Include numeric characters in the result. Default value is `true`. " +
					"**NOTE**: This is deprecated, use `numeric` instead.",
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.NumberNumericAttributePlanModifier(),
					planmodifiers.RequiresReplace(),
				},
				DeprecationMessage: "**NOTE**: This is deprecated, use `numeric` instead.",
			},

			"numeric": {
				Description: "Include numeric characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.NumberNumericAttributePlanModifier(),
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
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplaceIf(
						planmodifiers.RequiresReplaceUnlessEmptyStringToNull(),
						"Replace on modification unless updating from empty string (\"\") to null.",
						"Replace on modification unless updating from empty string (`\"\"`) to `null`.",
					),
				},
			},

			"result": {
				Description: "The generated random string.",
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},

			"id": {
				Description: "The generated random string.",
				Computed:    true,
				Type:        types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
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
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},

			"length": {
				Description: "The length of the string desired. The minimum value for length is 1 and, length " +
					"must also be >= (`min_upper` + `min_lower` + `min_numeric` + `min_special`).",
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
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
					resource.RequiresReplace(),
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

type stringModelV3 struct {
	ID              types.String `tfsdk:"id"`
	Keepers         types.Map    `tfsdk:"keepers"`
	Length          types.Int64  `tfsdk:"length"`
	Special         types.Bool   `tfsdk:"special"`
	Upper           types.Bool   `tfsdk:"upper"`
	Lower           types.Bool   `tfsdk:"lower"`
	Number          types.Bool   `tfsdk:"number"`
	Numeric         types.Bool   `tfsdk:"numeric"`
	MinNumeric      types.Int64  `tfsdk:"min_numeric"`
	MinUpper        types.Int64  `tfsdk:"min_upper"`
	MinLower        types.Int64  `tfsdk:"min_lower"`
	MinSpecial      types.Int64  `tfsdk:"min_special"`
	OverrideSpecial types.String `tfsdk:"override_special"`
	Result          types.String `tfsdk:"result"`
}
