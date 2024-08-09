// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AtLeastOneOfTrueValidator is the underlying struct implementing AtLeastOneOfTrue.
type AtLeastOneOfTrueValidator struct {
	PathExpressions path.Expressions
}

type AtLeastOneOfTrueValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    types.Bool
	Path           path.Path
	PathExpression path.Expression
}

type AtLeastOneOfTrueValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (av AtLeastOneOfTrueValidator) Description(ctx context.Context) string {
	return av.MarkdownDescription(ctx)
}

func (av AtLeastOneOfTrueValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that at least one attribute from this collection is set: %s", av.PathExpressions)
}

func (av AtLeastOneOfTrueValidator) Validate(ctx context.Context, req AtLeastOneOfTrueValidatorRequest, res *AtLeastOneOfTrueValidatorResponse) {
	if req.ConfigValue.IsNull() {
		return
	}

	// At least one value is not false
	if !req.ConfigValue.IsNull() && req.ConfigValue.ValueBool() {
		return
	}

	expressions := req.PathExpression.MergeExpressions(av.PathExpressions...)

	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)

		res.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			var mpVal types.Bool
			diags := req.Config.GetAttribute(ctx, mp, &mpVal)
			res.Diagnostics.Append(diags...)

			// Collect all errors
			if diags.HasError() {
				continue
			}

			// Delay validation until all involved attribute have a known value
			if mpVal.IsUnknown() {
				return
			}

			if mpVal.IsNull() {
				return
			}

			// At least one value is not false
			if !mpVal.IsNull() && mpVal.ValueBool() {
				return
			}
		}
	}

	expressions.Append(req.PathExpression)

	res.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
		req.Path,
		fmt.Sprintf("At least one attribute out of %s must be specified as true", expressions),
	))
}

func (av AtLeastOneOfTrueValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	validateReq := AtLeastOneOfTrueValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtLeastOneOfTrueValidatorResponse{}

	av.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func AtLeastOneOfTrue(expressions ...path.Expression) validator.Bool {
	return AtLeastOneOfTrueValidator{
		PathExpressions: expressions,
	}
}
