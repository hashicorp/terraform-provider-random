package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DefaultValue accepts an attr.Value and uses the supplied value to set a default if the config for
// the attribute is null.
func DefaultValue(val attr.Value) tfsdk.AttributePlanModifier {
	return &defaultValueAttributePlanModifier{val}
}

type defaultValueAttributePlanModifier struct {
	val attr.Value
}

func (d *defaultValueAttributePlanModifier) Description(ctx context.Context) string {
	return "If the config does not contain a value, a default will be set using val."
}

func (d *defaultValueAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

// Modify checks that the value of the attribute in the configuration and assigns the default value if
// the value in the config is null. This is a destructive operation in that it will overwrite any value
// present in the plan.
func (d *defaultValueAttributePlanModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	// Do not set default if the attribute configuration has been set.
	if !req.AttributeConfig.IsNull() {
		return
	}

	resp.AttributePlan = d.val
}

// RequiresReplace returns an attribute plan modifier that is identical to resource.RequiresReplace() with
// the exception that there is no check for `configRaw.IsNull && attrSchema.Computed` as a replacement
// needs to be triggered when the attribute has been removed from the config.
func RequiresReplace() tfsdk.AttributePlanModifier {
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

// Modify will trigger replacement (i.e., destroy-create) when `configRaw.IsNull && attrSchema.Computed`,
// which differs from the behaviour of `resource.RequiresReplace()`.
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

func NumberNumericAttributePlanModifier() tfsdk.AttributePlanModifier {
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

func (d *numberNumericAttributePlanModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
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

	if !numberConfig.Null && !numericConfig.Null && (numberConfig.Value != numericConfig.Value) {
		resp.Diagnostics.AddError(
			"Number and numeric are both configured with different values",
			"Number is deprecated, use numeric instead",
		)
		return
	}

	// Default to true for both number and numeric when both are null.
	if numberConfig.Null && numericConfig.Null {
		resp.AttributePlan = types.Bool{Value: true}
		return
	}

	// Default to using value for numeric if number is null.
	if numberConfig.IsNull() && !numericConfig.IsNull() {
		resp.AttributePlan = numericConfig
		return
	}

	// Default to using value for number if numeric is null.
	if !numberConfig.IsNull() && numericConfig.IsNull() {
		resp.AttributePlan = numberConfig
		return
	}
}

func RequiresReplaceIfValuesNotNull() tfsdk.AttributePlanModifier {
	return requiresReplaceIfValuesNotNullModifier{}
}

type requiresReplaceIfValuesNotNullModifier struct{}

func (r requiresReplaceIfValuesNotNullModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
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

	configMap, ok := req.AttributeConfig.(types.Map)
	if !ok {
		return
	}

	stateMap, ok := req.AttributeState.(types.Map)
	if !ok {
		return
	}

	replace := false

	for k, configValue := range configMap.Elems {
		stateValue, ok := stateMap.Elems[k]

		if !ok && configValue.IsNull() {
			continue
		}

		if !configValue.Equal(stateValue) {
			replace = true
			break
		}
	}

	if replace {
		resp.RequiresReplace = true
		return
	}

	//respPlan := resp.AttributePlan
	//
	//pm, ok := respPlan.(types.Map)
	//if ok {
	//	for k, v := range additionalElems {
	//		pm.Elems[k] = v
	//	}
	//}
	//
	//resp.AttributePlan = pm
}

// Description returns a human-readable description of the plan modifier.
func (r requiresReplaceIfValuesNotNullModifier) Description(ctx context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r requiresReplaceIfValuesNotNullModifier) MarkdownDescription(ctx context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}
