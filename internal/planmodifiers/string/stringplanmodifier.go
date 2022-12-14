package stringplanmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DefaultValue accepts a types.String value and uses the supplied value to set a default
// if the config for the attribute is null.
func DefaultValue(val types.String) planmodifier.String {
	return &defaultValueAttributePlanModifier{val}
}

type defaultValueAttributePlanModifier struct {
	val types.String
}

func (d *defaultValueAttributePlanModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If not configured, defaults to %s", d.val.ValueString())
}

func (d *defaultValueAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

// PlanModifyString checks that the value of the attribute in the configuration and assigns the default value if
// the value in the config is null. This is a destructive operation in that it will overwrite any value
// present in the plan.
func (d *defaultValueAttributePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do not set default if the attribute configuration has been set.
	if !req.ConfigValue.IsNull() {
		return
	}

	resp.PlanValue = d.val
}

// RequiresReplaceUnlessEmptyStringToNull returns a
// resource.RequiresReplaceIfFunc that returns true unless the change is from
// an empty string to null. This plan modifier allows practitioners to fix
// their configurations in this situation without replacing the resource.
//
// For example, version 3.4.2 errantly setup the random_password and
// random_string resource override_special attributes with an empty string
// default value plan modifier.
func RequiresReplaceUnlessEmptyStringToNull() stringplanmodifier.RequiresReplaceIfFunc {
	return func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
		// If the configuration is unknown, this cannot be sure what to do yet.
		if req.ConfigValue.IsUnknown() {
			resp.RequiresReplace = false
			return
		}

		// If the configuration is not null or the state is already null,
		// replace the resource.
		if !req.ConfigValue.IsNull() || req.StateValue.IsNull() {
			resp.RequiresReplace = true
			return
		}

		// If the state is not an empty string, replace the resource.
		if req.StateValue.ValueString() != "" {
			resp.RequiresReplace = true
			return
		}

		resp.RequiresReplace = false
	}
}
