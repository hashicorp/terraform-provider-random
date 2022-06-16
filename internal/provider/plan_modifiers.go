package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

func newNumberNumericAttributePlanModifier() tfsdk.AttributePlanModifier {
	return &numberNumericAttributePlanModifier{}
}

type numberNumericAttributePlanModifier struct {
}

func (d *numberNumericAttributePlanModifier) Description(ctx context.Context) string {
	return ""
}

func (d *numberNumericAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return ""
}

func (d *numberNumericAttributePlanModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	numberConfig := types.Bool{}
	diags := req.Config.GetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("number"), &numberConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	numericConfig := types.Bool{}
	req.Config.GetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("numeric"), &numericConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !numberConfig.Null && !numericConfig.Null {
		resp.Diagnostics.AddError(
			"Number numeric attribute plan modifier failed",
			"Cannot specify both number and numeric in config",
		)
		return
	}

	numberPlan := types.Bool{}
	diags = req.Config.GetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("number"), &numberPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	numericPlan := types.Bool{}
	req.Config.GetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("numeric"), &numericPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default to true for both number and numeric when both are null.
	if numberPlan.Null && numericPlan.Null {
		resp.AttributePlan = types.Bool{Value: true}
		return
	}

	// Default to using value for numeric if number is null
	if numberPlan.Null && !numericPlan.Null {
		resp.AttributePlan = numericPlan
		return
	}

	// Default to using value for number if numeric is null
	if !numberPlan.Null && numericPlan.Null {
		resp.AttributePlan = numberPlan
		return
	}
}

// RequiresReplace returns an attribute plan modifier that is identical to
// tfsdk.RequiresReplace() with the exception that there is no check for
// configRaw.IsNull && attrSchema.Computed because the defaults that we
// need to assign are computed.
func RequiresReplace() tfsdk.AttributePlanModifier {
	return RequiresReplaceModifier{}
}

type RequiresReplaceModifier struct{}

func (r RequiresReplaceModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeConfig == nil || req.AttributePlan == nil || req.AttributeState == nil {
		// shouldn't happen, but let's not panic if it does
		return
	}

	if req.State.Raw.IsNull() {
		// if we're creating the resource, no need to delete and
		// recreate it
		return
	}

	if req.Plan.Raw.IsNull() {
		// if we're deleting the resource, no need to delete and
		// recreate it
		return
	}

	if req.AttributePlan.Equal(req.AttributeState) {
		// if the plan and the state are in agreement, this attribute
		// isn't changing, don't require replace
		return
	}

	resp.RequiresReplace = true
}

// Description returns a human-readable description of the plan modifier.
func (r RequiresReplaceModifier) Description(ctx context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r RequiresReplaceModifier) MarkdownDescription(ctx context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}
