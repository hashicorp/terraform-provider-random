package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

type isAtLeastSumOfValidator struct {
	attributesToSum []*tftypes.AttributePath
}

var _ tfsdk.AttributeValidator = (*isAtLeastSumOfValidator)(nil)

func isAtLeastSumOf(attributePaths ...*tftypes.AttributePath) tfsdk.AttributeValidator {
	return &isAtLeastSumOfValidator{attributePaths}
}

func (av *isAtLeastSumOfValidator) Description(ctx context.Context) string {
	return av.MarkdownDescription(ctx)
}

func (av *isAtLeastSumOfValidator) MarkdownDescription(context.Context) string {
	return fmt.Sprintf("Ensure that attribute has a value >= sum of: %q", av.attributesToSum)
}

func (av *isAtLeastSumOfValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	tflog.Debug(ctx, "Validating that attribute has a value at least equal to the attributes to sum", map[string]interface{}{
		"attribute":       attrPathToString(req.AttributePath),
		"attributesToSum": av.attributesToSum,
	})

	var attrib types.Int64
	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, req.AttributeConfig, &attrib)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var sumOfAttribs int64
	var attributesToSumPaths []string

	for _, path := range av.attributesToSum {
		var attribToSum types.Int64
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path, &attribToSum)...)
		if resp.Diagnostics.HasError() {
			return
		}

		sumOfAttribs += attribToSum.Value
		attributesToSumPaths = append(attributesToSumPaths, attrPathToString(path))
	}

	if attrib.Value < sumOfAttribs {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			fmt.Sprintf("Attribute %q is less than summed attributes.", attrPathToString(req.AttributePath)),
			fmt.Sprintf(
				"Attribute %q (%d) cannot be less than %s (%d).",
				attrPathToString(req.AttributePath),
				attrib.Value,
				strings.Join(attributesToSumPaths, " + "),
				sumOfAttribs,
			),
		)
	}
}

// attrPathToString takes all the tftypes.AttributePathStep in a tftypes.AttributePath and concatenates them,
// using `.` as separator.
//
// This should be used only when trying to "print out" a tftypes.AttributePath in a log or an error message.
func attrPathToString(a *tftypes.AttributePath) string {
	var res strings.Builder
	for pos, step := range a.Steps() {
		if pos != 0 {
			res.WriteString(".")
		}
		switch v := step.(type) {
		case tftypes.AttributeName:
			res.WriteString(string(v))
		case tftypes.ElementKeyString:
			res.WriteString(string(v))
		case tftypes.ElementKeyInt:
			res.WriteString(strconv.FormatInt(int64(v), 10))
		case tftypes.ElementKeyValue:
			res.WriteString(tftypes.Value(v).String())
		}
	}
	return res.String()
}
