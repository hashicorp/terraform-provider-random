package int64planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DefaultValue accepts a types.Bool value and uses the supplied value to set a default
// if the config for the attribute is null.
func DefaultValue(val types.Int64) planmodifier.Int64 {
	return &defaultValueAttributePlanModifier{val}
}

type defaultValueAttributePlanModifier struct {
	val types.Int64
}

func (d *defaultValueAttributePlanModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If not configured, defaults to %d", d.val.ValueInt64())
}

func (d *defaultValueAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

// PlanModifyInt64 checks that the value of the attribute in the configuration and assigns the default value if
// the value in the config is null. This is a destructive operation in that it will overwrite any value
// present in the plan.
func (d *defaultValueAttributePlanModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// Do not set default if the attribute configuration has been set.
	if !req.ConfigValue.IsNull() {
		return
	}

	resp.PlanValue = d.val
}

// RequiresReplace returns an attribute plan modifier that is identical to resource.RequiresReplace() with
// the exception that there is no check for `configRaw.IsNull && attrSchema.Computed` as a replacement
// needs to be triggered when the attribute has been removed from the config.
func RequiresReplace() planmodifier.Int64 {
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

// PlanModifyInt64 will trigger replacement (i.e., destroy-create) when `configRaw.IsNull && attrSchema.Computed`,
// which differs from the behaviour of `resource.RequiresReplace()`.
func (r RequiresReplaceModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
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
