package provider_fm

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"provider not configured",
			"provider not configured",
		)
	}

	result, err := uuid.GenerateUUID()
	if err != nil {
		resp.Diagnostics.AddError(
			"error generating uuid",
			fmt.Sprintf("could not generate uuid: %s", err))
		return
	}

	u := &UUID{
		ID:     types.String{Value: result},
		Result: types.String{Value: result},
	}

	diags := resp.State.Set(ctx, u)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceUUID) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Intentionally left blank.
}

func (r resourceUUID) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	//TODO implement me
	panic("implement me")
}

func (r resourceUUID) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	resp.State.RemoveResource(ctx)
}

func (r resourceUUID) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)

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

	var state UUID

	state.ID.Value = result
	state.Result.Value = result

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
