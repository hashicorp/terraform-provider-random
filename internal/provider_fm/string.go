package provider_fm

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"math/big"
	"sort"
)

func getStringSchemaV1(sensitive bool, description string) tfsdk.Schema {
	idDesc := "The generated random string."
	if sensitive {
		idDesc = "A static value used internally by Terraform, this should not be referenced in configurations."
	}

	return tfsdk.Schema{
		Description: description,
		Attributes: map[string]tfsdk.Attribute{
			"keepers": {
				Description: "Arbitrary map of values that, when changed, will trigger recreation of " +
					"resource. See [the main provider documentation](../index.html) for more information.",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},

			"length": {
				Description: "The length of the string desired. The minimum value for length is 1 and, length " +
					"must also be >= (`min_upper` + `min_lower` + `min_numeric` + `min_special`).",
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
				Validators:    []tfsdk.AttributeValidator{validatorMinInt(1)},
			},

			"special": {
				Description: "Include special characters in the result. These are `!@#$%&*()-_=+[]{}<>:?`. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					defaultBool(true),
				},
			},

			"upper": {
				Description: "Include uppercase alphabet characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					defaultBool(true),
				},
			},

			"lower": {
				Description: "Include lowercase alphabet characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					defaultBool(true),
				},
			},

			"number": {
				Description: "Include numeric characters in the result. Default value is `true`.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					defaultBool(true),
				},
			},

			"min_numeric": {
				Description: "Minimum number of numeric characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					defaultInt(0),
				},
			},

			"min_upper": {
				Description: "Minimum number of uppercase alphabet characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					defaultInt(0),
				},
			},

			"min_lower": {
				Description: "Minimum number of lowercase alphabet characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					defaultInt(0),
				},
			},

			"min_special": {
				Description: "Minimum number of special characters in the result. Default value is `0`.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					defaultInt(0),
				},
			},

			"override_special": {
				Description: "Supply your own list of special characters to use for string generation.  This " +
					"overrides the default character list in the special argument.  The `special` argument must " +
					"still be set to true for any overwritten characters to be used in generation.",
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},

			"result": {
				Description: "The generated random string.",
				Type:        types.StringType,
				Computed:    true,
				Sensitive:   sensitive,
			},

			"id": {
				Description: idDesc,
				Computed:    true,
				Type:        types.StringType,
			},
		},
	}
}

func validatorMinInt(min int64) tfsdk.AttributeValidator {
	return minIntValidator{min}
}

type minIntValidator struct {
	val int64
}

func (m minIntValidator) Description(context.Context) string {
	return "MinInt validator ensures that attribute is at least val"
}

func (m minIntValidator) MarkdownDescription(context.Context) string {
	return "MinInt validator ensures that attribute is at least `val`"
}

func (m minIntValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	t := req.AttributeConfig.(types.Int64)

	if t.Value < m.val {
		resp.Diagnostics.AddError(
			fmt.Sprintf("expected attribute to be at least %d, got %d", m.val, t.Value),
			fmt.Sprintf("expected attribute to be at least %d, got %d", m.val, t.Value),
		)
	}
}

//nolint:unparam
func defaultBool(val bool) tfsdk.AttributePlanModifier {
	return boolDefault{val}
}

type boolDefault struct {
	val bool
}

func (d boolDefault) Description(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set using val."
}

func (d boolDefault) MarkdownDescription(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set using `val`."
}

func (d boolDefault) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	t := req.AttributeConfig.(types.Bool)

	if t.Null {
		resp.AttributePlan = types.Bool{
			Value: d.val,
		}
	}
}

func defaultInt(val int64) tfsdk.AttributePlanModifier {
	return intDefault{val}
}

type intDefault struct {
	val int64
}

func (d intDefault) Description(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set using val."
}

func (d intDefault) MarkdownDescription(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set using `val`."
}

func (d intDefault) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	t := req.AttributeConfig.(types.Int64)

	if t.Null {
		resp.AttributePlan = types.Int64{
			Null:  false,
			Value: d.val,
		}
	}
}

func defaultString(val string) tfsdk.AttributePlanModifier {
	return stringDefault{val}
}

type stringDefault struct {
	val string
}

func (d stringDefault) Description(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set."
}

func (d stringDefault) MarkdownDescription(ctx context.Context) string {
	return "If the plan does not contain a value, a default will be set."
}

func (d stringDefault) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	t := req.AttributeConfig.(types.String)

	if t.Null {
		resp.AttributePlan = types.String{
			Null:  false,
			Value: d.val,
		}
	}
}

func createString(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse, sensitive bool) {
	const numChars = "0123456789"
	const lowerChars = "abcdefghijklmnopqrstuvwxyz"
	const upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var specialChars = "!@#$%&*()-_=+[]{}<>:?"
	var plan StringModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	length := plan.Length.Value
	upper := plan.Upper.Value
	minUpper := plan.MinUpper.Value
	lower := plan.Lower.Value
	minLower := plan.MinLower.Value
	number := plan.Number.Value
	minNumeric := plan.MinNumeric.Value
	special := plan.Special.Value
	minSpecial := plan.MinSpecial.Value
	overrideSpecial := plan.OverrideSpecial.Value

	if overrideSpecial != "" {
		specialChars = overrideSpecial
	}

	var chars = string("")
	if upper {
		chars += upperChars
	}
	if lower {
		chars += lowerChars
	}
	if number {
		chars += numChars
	}
	if special {
		chars += specialChars
	}

	minMapping := map[string]int64{
		numChars:     minNumeric,
		lowerChars:   minLower,
		upperChars:   minUpper,
		specialChars: minSpecial,
	}

	var result = make([]byte, 0, length)

	for k, v := range minMapping {
		s, err := generateRandomBytes(&k, v)
		if err != nil {
			resp.Diagnostics.AddError(
				"error generating random bytes",
				fmt.Sprintf("error generating random bytes: %s", err),
			)
			return
		}
		result = append(result, s...)
	}

	s, err := generateRandomBytes(&chars, length-int64(len(result)))
	if err != nil {
		resp.Diagnostics.AddError(
			"error generating random bytes",
			fmt.Sprintf("error generating random bytes: %s", err),
		)
		return
	}

	result = append(result, s...)

	order := make([]byte, len(result))
	if _, err := rand.Read(order); err != nil {
		resp.Diagnostics.AddError(
			"error generating random bytes",
			fmt.Sprintf("error generating random bytes: %s", err),
		)
		return
	}

	sort.Slice(result, func(i, j int) bool {
		return order[i] < order[j]
	})

	str := StringModel{
		ID:              types.String{Value: string(result)},
		Keepers:         plan.Keepers,
		Length:          types.Int64{Value: length},
		Special:         types.Bool{Value: special},
		Upper:           types.Bool{Value: upper},
		Lower:           types.Bool{Value: lower},
		Number:          types.Bool{Value: number},
		MinNumeric:      types.Int64{Value: minNumeric},
		MinUpper:        types.Int64{Value: minUpper},
		MinLower:        types.Int64{Value: minLower},
		MinSpecial:      types.Int64{Value: minSpecial},
		OverrideSpecial: types.String{Value: overrideSpecial},
		Result:          types.String{Value: string(result)},
	}

	if sensitive {
		str.ID.Value = "none"
	}

	diags = resp.State.Set(ctx, str)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func generateRandomBytes(charSet *string, length int64) ([]byte, error) {
	bytes := make([]byte, length)
	setLen := big.NewInt(int64(len(*charSet)))
	for i := range bytes {
		idx, err := rand.Int(rand.Reader, setLen)
		if err != nil {
			return nil, err
		}
		bytes[i] = (*charSet)[idx.Int64()]
	}
	return bytes, nil
}

func importString(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse, sensitive bool) {
	id := req.ID

	state := StringModel{
		ID:     types.String{Value: id},
		Result: types.String{Value: id},
	}

	state.Keepers.ElemType = types.StringType

	if sensitive {
		state.ID.Value = "none"
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func validateLength(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
	var config StringModel
	req.Config.Get(ctx, &config)

	length := config.Length.Value
	minUpper := config.MinUpper.Value
	minLower := config.MinLower.Value
	minNumeric := config.MinNumeric.Value
	minSpecial := config.MinSpecial.Value

	if length < minUpper+minLower+minNumeric+minSpecial {
		resp.Diagnostics.AddError(
			fmt.Sprintf("length (%d) must be >= min_upper + min_lower + min_numeric + min_special (%d)", length, minUpper+minLower+minNumeric+minSpecial),
			fmt.Sprintf("length (%d) must be >= min_upper + min_lower + min_numeric + min_special (%d)", length, minUpper+minLower+minNumeric+minSpecial),
		)
	}
}
