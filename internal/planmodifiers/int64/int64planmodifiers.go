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
