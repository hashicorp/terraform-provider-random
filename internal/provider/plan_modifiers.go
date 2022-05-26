package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// newDefaultValueAttributePlanModifier accepts an attr.Value and returns a struct that implements the
// AttributePlanModifier interface.
//nolint:unparam // val is always true
func newDefaultValueAttributePlanModifier(val attr.Value) tfsdk.AttributePlanModifier {
	return &defaultValueAttributePlanModifier{val}
}

type defaultValueAttributePlanModifier struct {
	val attr.Value
}

func (d *defaultValueAttributePlanModifier) Description(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set using val."
}

func (d *defaultValueAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set using `val`."
}

// Modify checks that the value of the attribute in the configuration and the plan and only assigns the default value if
// the value in the config is null or the value in the plan is not known, or is known but is null.
func (d *defaultValueAttributePlanModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	// TODO: Remove once attr.Value interface includes IsNull.
	attribConfigValue, err := req.AttributeConfig.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Default value attribute plan modifier failed",
			fmt.Sprintf("Unable to convert attribute config (%s) to terraform value: %s", req.AttributeConfig.Type(ctx).String(), err),
		)
		return
	}

	// Do not set default if the attribute configuration has been set.
	if !attribConfigValue.IsNull() {
		return
	}

	// TODO: Remove once attr.Value interface includes IsUnknown.
	attribPlanValue, err := req.AttributePlan.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Default value attribute plan modifier failed",
			fmt.Sprintf("Unable to convert attribute plan %s to terraform value: %s", req.AttributePlan.Type(ctx).String(), err),
		)
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan modifier in the sequence has already been
	// applied and, we don't want to overwrite.
	if attribPlanValue.IsKnown() && !attribPlanValue.IsNull() {
		return
	}

	resp.AttributePlan = d.val
}
