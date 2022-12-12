package boolplanmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DefaultValue accepts a types.Bool value and uses the supplied value to set a default
// if the config for the attribute is null.
func DefaultValue(val types.Bool) planmodifier.Bool {
	return &defaultValueAttributePlanModifier{val}
}

type defaultValueAttributePlanModifier struct {
	val types.Bool
}

func (d *defaultValueAttributePlanModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If not configured, defaults to %t", d.val.ValueBool())
}

func (d *defaultValueAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

// PlanModifyBool checks that the value of the attribute in the configuration and assigns the default value if
// the value in the config is null. This is a destructive operation in that it will overwrite any value
// present in the plan.
func (d *defaultValueAttributePlanModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// Do not set default if the attribute configuration has been set.
	if !req.ConfigValue.IsNull() {
		return
	}

	resp.PlanValue = d.val
}

// RequiresReplace returns an attribute plan modifier that is identical to resource.RequiresReplace() with
// the exception that there is no check for `configRaw.IsNull && attrSchema.Computed` as a replacement
// needs to be triggered when the attribute has been removed from the config.
func RequiresReplace() planmodifier.Bool {
	return RequiresReplaceModifier{}
}

type RequiresReplaceModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (r RequiresReplaceModifier) Description(ctx context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r RequiresReplaceModifier) MarkdownDescription(ctx context.Context) string {
	return r.Description(ctx)
}

// PlanModifyBool will trigger replacement (i.e., destroy-create) when `configRaw.IsNull && attrSchema.Computed`,
// which differs from the behaviour of `resource.RequiresReplace()`.
func (r RequiresReplaceModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
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

	if req.PlanValue.Equal(req.StateValue) {
		// if the plan and the state are in agreement, this attribute
		// isn't changing, don't require replace
		return
	}

	resp.RequiresReplace = true
}

// NumberNumericAttributePlanModifier returns a plan modifier that keep the values
// held in number and numeric attributes synchronised.
func NumberNumericAttributePlanModifier() planmodifier.Bool {
	return &numberNumericAttributePlanModifier{}
}

type numberNumericAttributePlanModifier struct {
}

func (d *numberNumericAttributePlanModifier) Description(ctx context.Context) string {
	return "Ensures that number and numeric attributes are kept synchronised."
}

func (d *numberNumericAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d *numberNumericAttributePlanModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	numberConfig := types.Bool{}
	diags := req.Config.GetAttribute(ctx, path.Root("number"), &numberConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	numericConfig := types.Bool{}
	diags = req.Config.GetAttribute(ctx, path.Root("numeric"), &numericConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !numberConfig.IsNull() && !numericConfig.IsNull() && (numberConfig.ValueBool() != numericConfig.ValueBool()) {
		resp.Diagnostics.AddError(
			"Number and numeric are both configured with different values",
			"Number is deprecated, use numeric instead",
		)
		return
	}

	// Default to true for both number and numeric when both are null.
	if numberConfig.IsNull() && numericConfig.IsNull() {
		resp.PlanValue = types.BoolValue(true)
		return
	}

	// Default to using value for numeric if number is null.
	if numberConfig.IsNull() && !numericConfig.IsNull() {
		resp.PlanValue = numericConfig
		return
	}

	// Default to using value for number if numeric is null.
	if !numberConfig.IsNull() && numericConfig.IsNull() {
		resp.PlanValue = numberConfig
		return
	}
}
