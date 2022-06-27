package password

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/crypto/bcrypt"

	"github.com/terraform-providers/terraform-provider-random/internal/diagnostics"
	"github.com/terraform-providers/terraform-provider-random/internal/random"
)

var _ tfsdk.ResourceType = (*resourceType)(nil)

func NewResourceType() *resourceType {
	return &resourceType{}
}

type resourceType struct{}

func (r *resourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return schemaV2(), nil
}

func (r *resourceType) NewResource(_ context.Context, _ tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return &resource{}, nil
}

var (
	_ tfsdk.Resource                 = (*resource)(nil)
	_ tfsdk.ResourceWithImportState  = (*resource)(nil)
	_ tfsdk.ResourceWithUpgradeState = (*resource)(nil)
)

type resource struct{}

func (r *resource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan modelV2

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := random.RandomStringParams{
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

	result, err := random.CreateRandomString(params)
	if err != nil {
		resp.Diagnostics.Append(diagnostics.RandomReadError(err.Error())...)
		return
	}

	state := modelV2{
		ID:              types.String{Value: "none"},
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

	hash, err := generateHash(plan.Result.Value)
	if err != nil {
		resp.Diagnostics.Append(diagnostics.HashGenerationError(err.Error())...)
	}

	state.BcryptHash = types.String{Value: hash}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *resource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}

// Update is intentionally left blank as all required and optional attributes force replacement of the resource
// through the RequiresReplace AttributePlanModifier.
func (r *resource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *resource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

func (r *resource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	id := req.ID

	state := modelV2{
		ID:         types.String{Value: "none"},
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

	hash, err := generateHash(id)
	if err != nil {
		resp.Diagnostics.Append(diagnostics.HashGenerationError(err.Error())...)
	}

	state.BcryptHash = types.String{Value: hash}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resource) UpgradeState(context.Context) map[int64]tfsdk.ResourceStateUpgrader {
	schemaV0 := schemaV0()
	schemaV1 := schemaV1()

	return map[int64]tfsdk.ResourceStateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradePasswordStateV0toV2,
		},
		1: {
			PriorSchema:   &schemaV1,
			StateUpgrader: upgradePasswordStateV1toV2,
		},
	}
}

func upgradePasswordStateV0toV2(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
	type modelV0 struct {
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

	var passwordDataV0 modelV0

	resp.Diagnostics.Append(req.State.Get(ctx, &passwordDataV0)...)
	if resp.Diagnostics.HasError() {
		return
	}

	passwordDataV2 := modelV2{
		Keepers:         passwordDataV0.Keepers,
		Length:          passwordDataV0.Length,
		Special:         passwordDataV0.Special,
		Upper:           passwordDataV0.Upper,
		Lower:           passwordDataV0.Lower,
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
		resp.Diagnostics.Append(diagnostics.HashGenerationError(err.Error())...)
		return
	}

	passwordDataV2.BcryptHash.Value = hash

	diags := resp.State.Set(ctx, passwordDataV2)
	resp.Diagnostics.Append(diags...)
}

func upgradePasswordStateV1toV2(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
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
		BcryptHash      types.String `tfsdk:"bcrypt_hash"`
	}

	var passwordDataV1 modelV1

	resp.Diagnostics.Append(req.State.Get(ctx, &passwordDataV1)...)
	if resp.Diagnostics.HasError() {
		return
	}

	passwordDataV2 := modelV2{
		Keepers:         passwordDataV1.Keepers,
		Length:          passwordDataV1.Length,
		Special:         passwordDataV1.Special,
		Upper:           passwordDataV1.Upper,
		Lower:           passwordDataV1.Lower,
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

type modelV2 struct {
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
	BcryptHash      types.String `tfsdk:"bcrypt_hash"`
}
