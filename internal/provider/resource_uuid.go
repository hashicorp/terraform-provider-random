package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-random/internal/diagnostics"
	"github.com/terraform-providers/terraform-provider-random/internal/planmodifiers"
)

var (
	_ resource.Resource                = (*uuidResource)(nil)
	_ resource.ResourceWithImportState = (*uuidResource)(nil)
)

func NewUuidResource() resource.Resource {
	return &uuidResource{}
}

type uuidResource struct{}

func (r *uuidResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_uuid"
}

func (r *uuidResource) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.RequiresReplaceIfValuesNotNull(),
				},
			},
			"result": {
				Description: "The generated uuid presented in string format.",
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"id": {
				Description: "The generated uuid presented in string format.",
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
		},
	}, nil
}

func (r *uuidResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	result, err := uuid.GenerateUUID()
	if err != nil {
		resp.Diagnostics.AddError(
			"Create Random UUID error",
			"There was an error during generation of a UUID.\n\n"+
				diagnostics.RetryMsg+
				fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	var plan uuidModelV0

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	u := &uuidModelV0{
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

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *uuidResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update ensures the plan value is copied to the state to complete the update.
func (r *uuidResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model uuidModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *uuidResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *uuidResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	bytes, err := uuid.ParseUUID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Import Random UUID Error",
			"There was an error during the parsing of the UUID.\n\n"+
				diagnostics.RetryMsg+
				fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	result, err := uuid.FormatUUID(bytes)
	if err != nil {
		resp.Diagnostics.AddError(
			"Import Random UUID Error",
			"There was an error during the formatting of the UUID.\n\n"+
				diagnostics.RetryMsg+
				fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	var state uuidModelV0

	state.ID.Value = result
	state.Result.Value = result
	state.Keepers.ElemType = types.StringType

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

type uuidModelV0 struct {
	ID      types.String `tfsdk:"id"`
	Keepers types.Map    `tfsdk:"keepers"`
	Result  types.String `tfsdk:"result"`
}
