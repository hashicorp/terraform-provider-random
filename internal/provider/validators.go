package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// validatorMinInt accepts an int64 and returns a struct that implements the AttributeValidator interface.
func validatorMinInt(min int64) tfsdk.AttributeValidator {
	return minIntValidator{min}
}

type minIntValidator struct {
	val int64
}

func (m minIntValidator) Description(ctx context.Context) string {
	return "MinInt validator ensures that attribute is at least val"
}

func (m minIntValidator) MarkdownDescription(context.Context) string {
	return "MinInt validator ensures that attribute is at least `val`"
}

// Validate checks that the value of the attribute in the configuration is greater than or, equal to the value supplied
// when the minIntValidator struct was initialised.
func (m minIntValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var t types.Int64
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &t)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if t.Value < m.val {
		resp.Diagnostics.AddError(
			"Validating Min Int Error",
			fmt.Sprintf("Expected attribute at %s to be at least %d, got %d", req.AttributePath.String(), m.val, t.Value),
		)
	}
}