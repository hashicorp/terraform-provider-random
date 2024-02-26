// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	mapplanmodifiers "github.com/terraform-providers/terraform-provider-random/internal/planmodifiers/map"
	"github.com/terraform-providers/terraform-provider-random/internal/random"
)

var _ resource.Resource = (*shuffleResource)(nil)

func NewShuffleResource() resource.Resource {
	return &shuffleResource{}
}

type shuffleResource struct{}

func (r *shuffleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shuffle"
}

func (r *shuffleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `random_shuffle` generates a random permutation of a list of strings " +
			"given as an argument.",
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
			"seed": schema.StringAttribute{
				Description: "Arbitrary string with which to seed the random number generator, in order to " +
					"produce less-volatile permutations of the list.\n" +
					"\n" +
					"**Important:** Even with an identical seed, it is not guaranteed that the same permutation " +
					"will be produced across different versions of Terraform. This argument causes the " +
					"result to be *less volatile*, but not fixed for all time.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"input": schema.ListAttribute{
				Description: "The list of strings to shuffle.",
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"result_count": schema.Int64Attribute{
				Description: "The number of results to return. Defaults to the number of items in the " +
					"`input` list. If fewer items are requested, some elements will be excluded from the " +
					"result. If more items are requested, items will be repeated in the result but not more " +
					"frequently than the number of items in the input list.",
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"result": schema.ListAttribute{
				Description: "Random permutation of the list of strings given in `input`. The number of elements is determined by `result_count` if set, or the number of elements in `input`.",
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Description: "A static value used internally by Terraform, this should not be referenced in configurations.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *shuffleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data shuffleModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Legacy identifier attribute that is hardcoded. This is not necessary
	// after Terraform 0.12, but left for compatibility reasons. The attribute
	// could be removed in a future major version of the provider.
	data.ID = types.StringValue("-")

	inputElements := data.Input.Elements()

	var resultCount int64

	if !data.ResultCount.IsNull() {
		resultCount = data.ResultCount.ValueInt64()
	} else {
		resultCount = int64(len(inputElements))
	}

	// If the practitioner explicitly chose a result count of zero or the input
	// had no elements, immediately return with an empty list for the result.
	if resultCount == 0 || len(inputElements) == 0 {
		data.Result = types.ListValueMust(types.StringType, []attr.Value{})

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

		return
	}

	rand := random.NewRand(data.Seed.ValueString())
	resultElements := make([]attr.Value, 0, resultCount)

	// Keep producing permutations until we fill our result
Batches:
	for {
		perm := rand.Perm(len(inputElements))

		for _, i := range perm {
			resultElements = append(resultElements, inputElements[i])

			if int64(len(resultElements)) >= resultCount {
				break Batches
			}
		}
	}

	result, diags := types.ListValue(types.StringType, resultElements)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Result = result

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *shuffleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update ensures the plan value is copied to the state to complete the update.
func (r *shuffleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model shuffleModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *shuffleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

type shuffleModelV0 struct {
	ID          types.String `tfsdk:"id"`
	Keepers     types.Map    `tfsdk:"keepers"`
	Seed        types.String `tfsdk:"seed"`
	Input       types.List   `tfsdk:"input"`
	ResultCount types.Int64  `tfsdk:"result_count"`
	Result      types.List   `tfsdk:"result"`
}
