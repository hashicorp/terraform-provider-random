package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"golang.org/x/crypto/bcrypt"
)

type resourcePasswordType struct{}

func (r resourcePasswordType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return passwordSchemaV2(), nil
}

func (r resourcePasswordType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourcePassword{
		p: *(p.(*provider)),
	}, nil
}

type resourcePassword struct {
	p provider
}

func (r resourcePassword) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	createPassword(ctx, req, resp)
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r resourcePassword) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}

// Update is intentionally left blank as all required and optional attributes force replacement of the resource
// through the RequiresReplace AttributePlanModifier.
func (r resourcePassword) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r resourcePassword) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

func (r resourcePassword) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	importPassword(ctx, req, resp)
}

func (r resourcePassword) UpgradeState(context.Context) map[int64]tfsdk.ResourceStateUpgrader {
	schemaV0 := passwordSchemaV0()
	schemaV1 := passwordSchemaV1()

	return map[int64]tfsdk.ResourceStateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: migratePasswordStateV0toV2,
		},
		1: {
			PriorSchema:   &schemaV1,
			StateUpgrader: migratePasswordStateV1toV2,
		},
	}
}

func passwordSchemaV2() tfsdk.Schema {
	return tfsdk.Schema{
		Version: 2,
		Description: "Identical to [random_string](string.html) with the exception that the result is " +
			"treated as sensitive and, thus, _not_ displayed in console output. Read more about sensitive " +
			"data handling in the " +
			"[Terraform documentation](https://www.terraform.io/docs/language/state/sensitive-data.html).\n\n" +
			"This resource *does* use a cryptographic random number generator.",
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
					NewIntAtLeastValidator(1),
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
				Description: "Include numeric characters in the result. Default value is `true`. " +
					"**NOTE**: This is deprecated, use `numeric` instead.",
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					newNumberNumericAttributePlanModifier(),
				},
				DeprecationMessage: "**NOTE**: This is deprecated, use `numeric` instead.",
			},

			"numeric": {
				Description: "Include numeric characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					newNumberNumericAttributePlanModifier(),
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
				Sensitive:   true,
			},

			"bcrypt_hash": {
				Description: "A bcrypt hash of the generated random string.",
				Type:        types.StringType,
				Computed:    true,
				Sensitive:   true,
			},

			"id": {
				Description: "A static value used internally by Terraform, this should not be referenced in configurations.",
				Computed:    true,
				Type:        types.StringType,
			},
		},
	}
}

func passwordSchemaV1() tfsdk.Schema {
	return tfsdk.Schema{
		Version: 1,
		Description: "Identical to [random_string](string.html) with the exception that the result is " +
			"treated as sensitive and, thus, _not_ displayed in console output. Read more about sensitive " +
			"data handling in the " +
			"[Terraform documentation](https://www.terraform.io/docs/language/state/sensitive-data.html).\n\n" +
			"This resource *does* use a cryptographic random number generator.",
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
					NewIntAtLeastValidator(1),
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
				Sensitive:   true,
			},

			"bcrypt_hash": {
				Description: "A bcrypt hash of the generated random string.",
				Type:        types.StringType,
				Computed:    true,
				Sensitive:   true,
			},

			"id": {
				Description: "A static value used internally by Terraform, this should not be referenced in configurations.",
				Computed:    true,
				Type:        types.StringType,
			},
		},
	}
}

func passwordSchemaV0() tfsdk.Schema {
	return tfsdk.Schema{
		Description: "Identical to [random_string](string.html) with the exception that the result is " +
			"treated as sensitive and, thus, _not_ displayed in console output. Read more about sensitive " +
			"data handling in the " +
			"[Terraform documentation](https://www.terraform.io/docs/language/state/sensitive-data.html).\n\n" +
			"This resource *does* use a cryptographic random number generator.",
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
					NewIntAtLeastValidator(1),
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
				Sensitive:   true,
			},

			"id": {
				Description: "A static value used internally by Terraform, this should not be referenced in configurations.",
				Computed:    true,
				Type:        types.StringType,
			},
		},
	}
}

// createPassword currently uses plan.Number.Value for both number and numeric.
// TODO: Remove usage of number before release.
func createPassword(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan PasswordModelV2

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

	state := PasswordModelV2{
		ID:              types.String{Value: "none"},
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

	hash, err := generateHash(plan.Result.Value)
	if err != nil {
		resp.Diagnostics.Append(hashGenerationError(err.Error())...)
	}

	state.BcryptHash = types.String{Value: hash}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func importPassword(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	id := req.ID

	state := PasswordModelV2{
		ID:     types.String{Value: "none"},
		Result: types.String{Value: id},
	}

	state.Keepers.ElemType = types.StringType

	hash, err := generateHash(id)
	if err != nil {
		resp.Diagnostics.Append(hashGenerationError(err.Error())...)
	}

	state.BcryptHash = types.String{Value: hash}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func migratePasswordStateV0toV2(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
	var passwordDataV0 PasswordModelV0

	resp.Diagnostics.Append(req.State.Get(ctx, &passwordDataV0)...)
	if resp.Diagnostics.HasError() {
		return
	}

	passwordDataV2 := PasswordModelV2{
		Keepers:         passwordDataV0.Keepers,
		Length:          passwordDataV0.Length,
		Special:         passwordDataV0.Special,
		Upper:           passwordDataV0.Upper,
		Lower:           passwordDataV0.Lower,
		Number:          passwordDataV0.Number,
		Numeric:         passwordDataV0.Number,
		MinNumeric:      passwordDataV0.MinNumeric,
		MinLower:        passwordDataV0.MinLower,
		MinSpecial:      passwordDataV0.MinSpecial,
		OverrideSpecial: passwordDataV0.OverrideSpecial,
		Result:          passwordDataV0.Result,
		ID:              passwordDataV0.ID,
	}

	hash, err := generateHash(passwordDataV2.Result.Value)
	if err != nil {
		resp.Diagnostics.Append(hashGenerationError(err.Error())...)
		return
	}

	passwordDataV2.BcryptHash.Value = hash

	diags := resp.State.Set(ctx, passwordDataV2)
	resp.Diagnostics.Append(diags...)
}

func migratePasswordStateV1toV2(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
	var passwordDataV1 PasswordModelV1

	resp.Diagnostics.Append(req.State.Get(ctx, &passwordDataV1)...)
	if resp.Diagnostics.HasError() {
		return
	}

	passwordDataV2 := PasswordModelV2{
		Keepers:         passwordDataV1.Keepers,
		Length:          passwordDataV1.Length,
		Special:         passwordDataV1.Special,
		Upper:           passwordDataV1.Upper,
		Lower:           passwordDataV1.Lower,
		Number:          passwordDataV1.Number,
		Numeric:         passwordDataV1.Number,
		MinNumeric:      passwordDataV1.MinNumeric,
		MinLower:        passwordDataV1.MinLower,
		MinSpecial:      passwordDataV1.MinSpecial,
		OverrideSpecial: passwordDataV1.OverrideSpecial,
		BcryptHash:      passwordDataV1.BcryptHash,
		Result:          passwordDataV1.Result,
		ID:              passwordDataV1.ID,
	}

	diags := resp.State.Set(ctx, passwordDataV2)
	resp.Diagnostics.Append(diags...)
}

func generateHash(toHash string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(toHash), bcrypt.DefaultCost)

	return string(hash), err
}
