package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

func (r resourcePassword) ValidateConfig(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
	validateLength(ctx, req, resp)
}

func (r resourcePassword) UpgradeState(context.Context) map[int64]tfsdk.ResourceStateUpgrader {
	return map[int64]tfsdk.ResourceStateUpgrader{
		0: {
			StateUpgrader: migratePasswordStateV0toV1,
		},
	}
}

func getPasswordSchemaV1() tfsdk.Schema {
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

	passwordSchema.Attributes["bcrypt_hash"] = tfsdk.Attribute{
		Description: "A bcrypt hash of the generated random string.",
		Type:        types.StringType,
		Computed:    true,
		Sensitive:   true,
	}

	passwordSchema.Version = 1

	return passwordSchema
}

func createPassword(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan PasswordModel

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

	state := PasswordModel{
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

	state := PasswordModel{
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

func passwordDataTftypesV0() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"keepers": tftypes.Map{
				ElementType: tftypes.String,
			},
			"length":           tftypes.Number,
			"special":          tftypes.Bool,
			"upper":            tftypes.Bool,
			"lower":            tftypes.Bool,
			"number":           tftypes.Bool,
			"min_numeric":      tftypes.Number,
			"min_upper":        tftypes.Number,
			"min_lower":        tftypes.Number,
			"min_special":      tftypes.Number,
			"override_special": tftypes.String,
			"result":           tftypes.String,
			"id":               tftypes.String,
		},
	}
}

func passwordDataTftypesV1() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"bcrypt_hash": tftypes.String,
			"keepers": tftypes.Map{
				ElementType: tftypes.String,
			},
			"length":           tftypes.Number,
			"special":          tftypes.Bool,
			"upper":            tftypes.Bool,
			"lower":            tftypes.Bool,
			"number":           tftypes.Bool,
			"min_numeric":      tftypes.Number,
			"min_upper":        tftypes.Number,
			"min_lower":        tftypes.Number,
			"min_special":      tftypes.Number,
			"override_special": tftypes.String,
			"result":           tftypes.String,
			"id":               tftypes.String,
		},
	}
}

func migratePasswordStateV0toV1(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
	rawStateValue, err := req.RawState.Unmarshal(passwordDataTftypesV0())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to migrate password state from v0 to v1",
			fmt.Sprintf("Unmarshalling prior state failed: %s", err),
		)
		return
	}

	var rawState map[string]tftypes.Value
	if err := rawStateValue.As(&rawState); err != nil {
		resp.Diagnostics.AddError(
			"Unable to migrate password state from v0 to v1",
			fmt.Sprintf("Unable to convert prior state: %s", err),
		)
		return
	}

	if _, ok := rawState["result"]; !ok {
		resp.Diagnostics.AddError(
			"Unable to migrate password state from v0 to v1",
			"Prior state does not contain result.",
		)
		return
	}

	var result string
	if err := rawState["result"].As(&result); err != nil {
		resp.Diagnostics.AddError(
			"Unable to migrate password state from v0 to v1",
			"Result from prior state could not be converted to string.\n\n"+
				"As result is a sensitive value you will need to inspect it by looking\n"+
				"at the state file directly as 'terraform state show' will display (sensitive value).",
		)
		return
	}

	hash, err := generateHash(result)
	if err != nil {
		resp.Diagnostics.Append(hashGenerationError(err.Error())...)
		return
	}

	dynamicValue, err := tfprotov6.NewDynamicValue(
		passwordDataTftypesV1(),
		tftypes.NewValue(passwordDataTftypesV1(), map[string]tftypes.Value{
			"bcrypt_hash":      tftypes.NewValue(tftypes.String, hash),
			"keepers":          rawState["keepers"],
			"length":           rawState["length"],
			"special":          rawState["special"],
			"upper":            rawState["upper"],
			"lower":            rawState["lower"],
			"number":           rawState["number"],
			"min_numeric":      rawState["min_numeric"],
			"min_upper":        rawState["min_upper"],
			"min_lower":        rawState["min_lower"],
			"min_special":      rawState["min_special"],
			"override_special": rawState["override_special"],
			"result":           rawState["result"],
			"id":               rawState["id"],
		}))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to migrate password state from v0 to v1",
			fmt.Sprintf("Failed to generate new dynamic value: %s", err),
		)
		return
	}

	resp.DynamicValue = &dynamicValue
}

func generateHash(toHash string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(toHash), bcrypt.DefaultCost)

	return string(hash), err
}
