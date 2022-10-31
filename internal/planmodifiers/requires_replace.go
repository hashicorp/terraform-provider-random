package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RequiresReplaceUnlessEmptyStringToNull returns a
// resource.RequiresReplaceIfFunc that returns true unless the change is from
// an empty string to null. This plan modifier allows practitioners to fix
// their configurations in this situation without replacing the resource.
//
// For example, version 3.4.2 errantly setup the random_password and
// random_string resource override_special attributes with an empty string
// default value plan modifier.
func RequiresReplaceUnlessEmptyStringToNull() resource.RequiresReplaceIfFunc {
	return func(ctx context.Context, state attr.Value, config attr.Value, path path.Path) (bool, diag.Diagnostics) {
		// If the configuration is unknown, this cannot be sure what to do yet.
		if config.IsUnknown() {
			return false, nil
		}

		// If the configuration is not null or the state is already null,
		// replace the resource.
		if !config.IsNull() || state.IsNull() {
			return true, nil
		}

		// If the state is not a string value, return an error diagnostic
		// about the errant implementation.
		var stateString types.String

		diags := tfsdk.ValueAs(ctx, state, &stateString)

		if diags.HasError() {
			return false, diags
		}

		// If the state is not an empty string, replace the resource.
		if stateString.ValueString() != "" {
			return true, diags
		}

		return false, diags
	}
}
