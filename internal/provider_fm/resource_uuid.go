package provider_fm

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceUUIDType struct{}

func (r resourceUUIDType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "The resource `random_uuid` generates random uuid string that is intended to be " +
			"used as unique identifiers for other resources.\n" +
			"\n" +
			"This resource uses [hashicorp/go-uuid](https://github.com/hashicorp/go-uuid) to generate a " +
			"UUID-formatted string for use with services needed a unique string identifier.",
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
			"result": {
				Description: "The generated uuid presented in string format.",
				Type:        types.StringType,
				Computed:    true,
			},
			"id": {
				Description: "The generated uuid presented in string format.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (r resourceUUIDType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceUUID{
		p: *(p.(*provider)),
	}, nil
}

type resourceUUID struct {
	p provider
}

func (r resourceUUID) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	result, err := uuid.GenerateUUID()
	if err != nil {
		resp.Diagnostics.AddError(
			"error generating uuid",
			fmt.Sprintf("could not generate uuid: %s", err))
		return
	}

	var plan UUIDModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	u := &UUIDModel{
		ID:      types.String{Value: result},
		Result:  types.String{Value: result},
		Keepers: plan.Keepers,
	}

	diags = resp.State.Set(ctx, u)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceUUID) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Intentionally left blank.
}

func (r resourceUUID) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Intentionally left blank.
}

func (r resourceUUID) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	resp.State.RemoveResource(ctx)
}

func (r resourceUUID) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	bytes, err := uuid.ParseUUID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"error parsing uuid bytes",
			fmt.Sprintf("error parsing uuid bytes: %s", err))
		return
	}

	result, err := uuid.FormatUUID(bytes)
	if err != nil {
		resp.Diagnostics.AddError(
			"error formatting uuid bytes",
			fmt.Sprintf("error formatting uuid bytes: %s", err))
		return
	}

	var state UUIDModel

	state.ID.Value = result
	state.Result.Value = result
	state.Keepers.ElemType = types.StringType

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
