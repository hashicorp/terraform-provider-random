// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-random/internal/diagnostics"
	mapplanmodifiers "github.com/terraform-providers/terraform-provider-random/internal/planmodifiers/map"
)

var (
	_ resource.Resource                = (*uuidV7Resource)(nil)
	_ resource.ResourceWithImportState = (*uuidV7Resource)(nil)
)

func NewUuidV7Resource() resource.Resource {
	return &uuidV7Resource{}
}

type uuidV7Resource struct{}

func (r *uuidV7Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_uuid7"
}

func (r *uuidV7Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `random_uuid7` generates a random version 7 uuid string that is intended " +
			"to be used as a unique identifier for other resources.\n" +
			"\n" +
			"This resource uses [google/uuid](https://github.com/google/uuid) to generate a " +
			"valid V7 UUID for use with services needing a unique string identifier.",
		Attributes: map[string]schema.Attribute{
			"keepers": schema.MapAttribute{
				Description: "Arbitrary map of values that, when changed, will trigger recreation of " +
					"resource. See [the main provider documentation](../index.html) for more information.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifiers.RequiresReplaceIfValuesNotNull(),
				},
			},
			"result": schema.StringAttribute{
				Description: "The generated uuid presented in string format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The generated uuid presented in string format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *uuidV7Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	result, err := uuid.NewV7()
	if err != nil {
		resp.Diagnostics.AddError(
			"Create Random UUID v7 error",
			"There was an error during generation of a UUID.\n\n"+
				diagnostics.RetryMsg+
				fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	var plan uuidModelV7

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	u := &uuidModelV7{
		ID:      types.StringValue(result.String()),
		Result:  types.StringValue(result.String()),
		Keepers: plan.Keepers,
	}

	diags = resp.State.Set(ctx, u)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *uuidV7Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update ensures the plan value is copied to the state to complete the update.
func (r *uuidV7Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model uuidModelV7

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *uuidV7Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *uuidV7Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parsedUuid, err := uuid.Parse(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Import Random UUID Error",
			"There was an error during the parsing of the UUID.\n\n"+
				diagnostics.RetryMsg+
				fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	var state uuidModelV7

	state.ID = types.StringValue(parsedUuid.String())
	state.Result = types.StringValue(parsedUuid.String())
	state.Keepers = types.MapNull(types.StringType)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

type uuidModelV7 struct {
	ID      types.String `tfsdk:"id"`
	Keepers types.Map    `tfsdk:"keepers"`
	Result  types.String `tfsdk:"result"`
}
