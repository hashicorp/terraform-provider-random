package provider_fm

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
	"strings"
)

type resourceIntegerType struct{}

func (r resourceIntegerType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "The resource `random_integer` generates random values from a given range, described " +
			"by the `min` and `max` attributes of a given resource.\n" +
			"\n" +
			"This resource can be used in conjunction with resources that have the `create_before_destroy` " +
			"lifecycle flag set, to avoid conflicts with unique names during the brief period where both the " +
			"old and new resources exist concurrently.",
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
			"min": {
				Description:   "The minimum inclusive value of the range.",
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"max": {
				Description:   "The maximum inclusive value of the range.",
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"seed": {
				Description:   "A custom seed to always produce the same value.",
				Type:          types.StringType,
				Optional:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"result": {
				Description: "The random integer result.",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"id": {
				Description: "The string representation of the integer result.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (r resourceIntegerType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceInteger{
		p: *(p.(*provider)),
	}, nil
}

type resourceInteger struct {
	p provider
}

func (r resourceInteger) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan IntegerModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	max := int(plan.Max.Value)
	min := int(plan.Min.Value)
	seed := plan.Seed.Value

	if max < min {
		resp.Diagnostics.AddError(
			"minimum value needs to be smaller than or equal to maximum value",
			"minimum value needs to be smaller than or equal to maximum value",
		)
		return
	}

	rand := NewRand(seed)
	number := rand.Intn((max+1)-min) + min

	u := &IntegerModel{
		ID:      types.String{Value: strconv.Itoa(number)},
		Keepers: plan.Keepers,
		Min:     types.Int64{Value: int64(min)},
		Max:     types.Int64{Value: int64(max)},
		Result:  types.Int64{Value: int64(number)},
	}

	if seed != "" {
		u.Seed.Value = seed
	} else {
		u.Seed.Null = true
	}

	diags = resp.State.Set(ctx, u)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceInteger) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Intentionally left blank.
}

func (r resourceInteger) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Intentionally left blank.
}

func (r resourceInteger) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	resp.State.RemoveResource(ctx)
}

func (r resourceInteger) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	parts := strings.Split(req.ID, ",")
	if len(parts) != 3 && len(parts) != 4 {
		resp.Diagnostics.AddError(
			"Invalid import usage: expecting {result},{min},{max} or {result},{min},{max},{seed}",
			"Invalid import usage: expecting {result},{min},{max} or {result},{min},{max},{seed}",
		)
		return
	}

	result, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"error parsing result",
			fmt.Sprintf("error parsing result: %s", err),
		)
		return
	}

	min, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"error parsing min",
			fmt.Sprintf("error parsing min: %s", err),
		)
		return
	}

	max, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"error parsing max",
			fmt.Sprintf("error parsing max: %s", err),
		)
		return
	}

	var state IntegerModel

	state.ID.Value = parts[0]
	state.Keepers.ElemType = types.StringType
	state.Result.Value = result
	state.Min.Value = min
	state.Max.Value = max

	if len(parts) == 4 {
		state.Seed.Value = parts[3]
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
