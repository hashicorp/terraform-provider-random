package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/crypto/bcrypt"
)

type resourcePasswordType struct{}

func (r resourcePasswordType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return getPasswordSchemaV1(), nil
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
	passwordSchemaV0 := getPasswordSchemaV0()

	return map[int64]tfsdk.ResourceStateUpgrader{
		0: {
			PriorSchema:   &passwordSchemaV0,
			StateUpgrader: migratePasswordStateV0toV1,
		},
	}
}

func getPasswordSchemaV1() tfsdk.Schema {
	passwordSchema := getPasswordSchemaV0()

	passwordSchema.Attributes["bcrypt_hash"] = tfsdk.Attribute{
		Description: "A bcrypt hash of the generated random string.",
		Type:        types.StringType,
		Computed:    true,
		Sensitive:   true,
	}

	passwordSchema.Version = 1

	return passwordSchema
}

func getPasswordSchemaV0() tfsdk.Schema {
	passwordSchema := passwordStringSchema()

	passwordSchema.Description = "Identical to [random_string](string.html) with the exception that the result is " +
		"treated as sensitive and, thus, _not_ displayed in console output. Read more about sensitive " +
		"data handling in the " +
		"[Terraform documentation](https://www.terraform.io/docs/language/state/sensitive-data.html).\n\n" +
		"This resource *does* use a cryptographic random number generator."

	id, ok := passwordSchema.Attributes["id"]
	if ok {
		id.Description = "A static value used internally by Terraform, this should not be referenced in configurations."
		passwordSchema.Attributes["id"] = id
	}

	result, ok := passwordSchema.Attributes["result"]
	if ok {
		result.Sensitive = true
		passwordSchema.Attributes["result"] = result
	}

	return passwordSchema
}

func createPassword(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan PasswordModelV1

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
		number:          plan.Number.Value,
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

	state := PasswordModelV1{
		ID:              types.String{Value: "none"},
		Keepers:         plan.Keepers,
		Length:          types.Int64{Value: plan.Length.Value},
		Special:         types.Bool{Value: plan.Special.Value},
		Upper:           types.Bool{Value: plan.Upper.Value},
		Lower:           types.Bool{Value: plan.Lower.Value},
		Number:          types.Bool{Value: plan.Number.Value},
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

	state := PasswordModelV1{
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

func migratePasswordStateV0toV1(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
	var passwordDataV0 PasswordModelV0

	resp.Diagnostics.Append(req.State.Get(ctx, &passwordDataV0)...)
	if resp.Diagnostics.HasError() {
		return
	}

	passwordDataV1 := PasswordModelV1{
		Keepers:         passwordDataV0.Keepers,
		Length:          passwordDataV0.Length,
		Special:         passwordDataV0.Special,
		Upper:           passwordDataV0.Upper,
		Lower:           passwordDataV0.Lower,
		Number:          passwordDataV0.Number,
		MinNumeric:      passwordDataV0.MinNumeric,
		MinLower:        passwordDataV0.MinLower,
		MinSpecial:      passwordDataV0.MinSpecial,
		OverrideSpecial: passwordDataV0.OverrideSpecial,
		Result:          passwordDataV0.Result,
		ID:              passwordDataV0.ID,
	}

	hash, err := generateHash(passwordDataV1.Result.Value)
	if err != nil {
		resp.Diagnostics.Append(hashGenerationError(err.Error())...)
		return
	}

	passwordDataV1.BcryptHash.Value = hash

	resp.Diagnostics.Append(resp.State.Set(ctx, passwordDataV1)...)
}

func generateHash(toHash string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(toHash), bcrypt.DefaultCost)

	return string(hash), err
}
