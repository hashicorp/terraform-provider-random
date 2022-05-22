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

// intAtLeastValidator checks that the value of the attribute in the configuration
// (i.e., AttributeConfig in ValidateAttributeRequest) is greater than or, equal to minVal.
type intAtLeastValidator struct {
	minVal int64
}

var _ tfsdk.AttributeValidator = (*intAtLeastValidator)(nil)

func NewIntAtLeastValidator(min int64) tfsdk.AttributeValidator {
	return &intAtLeastValidator{min}
}

func (av *intAtLeastValidator) Description(ctx context.Context) string {
	return "intAtLeastValidator ensures that attribute is at least minVal"
}

func (av *intAtLeastValidator) MarkdownDescription(context.Context) string {
	return "intAtLeastValidator ensures that attribute is at least `minVal`"
}

// Validate first determines whether AttributeConfig (attr.Value) can be reflected into types.Int64 and then checks
// that the value is >= minVal.
func (av *intAtLeastValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var t types.Int64

	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, req.AttributeConfig, &t)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if t.Value < av.minVal {
		resp.Diagnostics.AddError(
			"Validating Int At Least Error",
			fmt.Sprintf("Attribute %q (%d) must be at least %d", attrPathToString(req.AttributePath), t.Value, av.minVal),
		)
	}
}

// intIsAtLeastValidator checks that the value of the attribute in the configuration
// (i.e., AttributeConfig in ValidateAttributeRequest) is greater than or, equal to the sum of the values of the
// attributes in the slice of AttributePath.
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

// Validate first determines whether AttributeConfig (attr.Value) can be reflected into types.Int64 and then checks
// that the value is >= sum of values of the attributes defined in attributesToSum.
func (av *intIsAtLeastSumOfValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
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
