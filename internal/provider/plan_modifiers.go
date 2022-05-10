package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// defaultBool accepts a bool and returns a struct that implements the AttributePlanModifier interface.
//nolint:unparam // val is always true
func defaultBool(val bool) tfsdk.AttributePlanModifier {
	return boolDefault{val}
}

type boolDefault struct {
	val bool
}

func (d boolDefault) Description(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set using val."
}

func (d boolDefault) MarkdownDescription(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set using `val`."
}

// Modify checks that the value of the attribute in the configuration and, if the attribute is Null, assigns the value
// supplied to the boolDefault struct when it was initialised.
func (d boolDefault) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	var t types.Bool
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &t)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if t.Null {
		resp.AttributePlan = types.Bool{
			Value: d.val,
		}
	}
}

// defaultInt accepts an int64 and returns a struct that implements the AttributePlanModifier interface.
func defaultInt(val int64) tfsdk.AttributePlanModifier {
	return intDefault{val}
}

type intDefault struct {
	val int64
}

func (d intDefault) Description(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set using val."
}

func (d intDefault) MarkdownDescription(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set using `val`."
}

// Modify checks that the value of the attribute in the configuration and, if the attribute is Null, assigns the value
// supplied to the intDefault struct when it was initialised.
func (d intDefault) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	var t types.Int64
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &t)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if t.Null {
		resp.AttributePlan = types.Int64{
			Null:  false,
			Value: d.val,
		}
	}
}

// defaultString accepts a string and returns a struct that implements the AttributePlanModifier interface.
func defaultString(val string) tfsdk.AttributePlanModifier {
	return stringDefault{val}
}

type stringDefault struct {
	val string
}

func (d stringDefault) Description(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set."
}

func (d stringDefault) MarkdownDescription(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set."
}

// Modify checks that the value of the attribute in the configuration and, if the attribute is Null, assigns the value
// supplied to the stringDefault struct when it was initialised.
func (d stringDefault) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	var t types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &t)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if t.Null {
		resp.AttributePlan = types.String{
			Null:  false,
			Value: d.val,
		}
	}
}
