package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceStringType struct{}

func (r resourceStringType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return getStringSchemaV1(), nil
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

func getStringSchemaV1() tfsdk.Schema {
	stringSchema := passwordStringSchema()

	stringSchema.Description = "The resource `random_string` generates a random permutation of alphanumeric " +
		"characters and optionally special characters.\n" +
		"\n" +
		"This resource *does* use a cryptographic random number generator.\n" +
		"\n" +
		"Historically this resource's intended usage has been ambiguous as the original example used " +
		"it in a password. For backwards compatibility it will continue to exist. For unique ids please " +
		"use [random_id](id.html), for sensitive random values please use [random_password](password.html)."

	id, ok := stringSchema.Attributes["id"]
	if ok {
		id.Description = "The generated random string."
		stringSchema.Attributes["id"] = id
	}

	stringSchema.Attributes["number"] = tfsdk.Attribute{
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
	}

	stringSchema.Attributes["numeric"] = tfsdk.Attribute{
		Description: "Include numeric characters in the result. Default value is `true`.",
		Type:        types.BoolType,
		Optional:    true,
		Computed:    true,
		PlanModifiers: []tfsdk.AttributePlanModifier{
			tfsdk.RequiresReplace(),
			newNumberNumericAttributePlanModifier(),
		},
	}

	return stringSchema
}

func createString(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan StringModelV1

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

	state := StringModelV1{
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

	state := StringModelV0{
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
