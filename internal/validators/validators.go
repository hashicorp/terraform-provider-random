package validators

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

// intIsAtLeastSumOfValidator checks that the value of the attribute in the configuration
// (i.e., AttributeConfig in ValidateAttributeRequest) is greater than or, equal to the sum of the values of the
// attributes in the slice of AttributePath.
// TODO: Remove once https://github.com/hashicorp/terraform-plugin-framework-validators/pull/29 is merged.
type intIsAtLeastSumOfValidator struct {
	attributesToSum []*tftypes.AttributePath
}

var _ tfsdk.AttributeValidator = (*intIsAtLeastSumOfValidator)(nil)

func NewIntIsAtLeastSumOfValidator(attributePaths ...*tftypes.AttributePath) tfsdk.AttributeValidator {
	return &intIsAtLeastSumOfValidator{attributePaths}
}

func (av *intIsAtLeastSumOfValidator) Description(ctx context.Context) string {
	return av.MarkdownDescription(ctx)
}

func (av *intIsAtLeastSumOfValidator) MarkdownDescription(context.Context) string {
	return fmt.Sprintf("Ensure that attribute has a value >= sum of: %q", av.attributesToSum)
}

// Validate runs the following checks:
// 1. Determines if AttributeConfig can be reflected into types.Int64.
// 2. Checks that the AttributeConfig value is >= sum of values of the attributes defined in attributesToSum.
func (av *intIsAtLeastSumOfValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	tflog.Debug(ctx, "Validating that attribute has a value at least equal to the attributes to sum", map[string]interface{}{
		"attribute":       attrPathToString(req.AttributePath),
		"attributesToSum": av.attributesToSum,
	})

	// TODO: Remove once attr.Value interface includes IsNull.
	attribConfigValue, err := req.AttributeConfig.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Int at least sum of validator failed",
			fmt.Sprintf("Unable to convert attribute config (%s) to terraform value: %s", req.AttributeConfig.Type(ctx).String(), err),
		)
		return
	}

	if attribConfigValue.IsNull() || !attribConfigValue.IsKnown() {
		return
	}

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
		attribPath := attrPathToString(req.AttributePath)

		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			fmt.Sprintf("Attribute %q is less than summed attributes.", attribPath),
			fmt.Sprintf("Attribute %q (%d) cannot be less than %s (%d).", attribPath, attrib.Value, strings.Join(attributesToSumPaths, " + "), sumOfAttribs),
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
