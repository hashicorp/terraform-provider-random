package pet

import (
	"context"
	"fmt"
	"strings"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-random/internal/planmodifiers"
)

func NewResourceType() *resourceType {
	return &resourceType{}
}

var _ tfsdk.ResourceType = (*resourceType)(nil)

type resourceType struct{}

func (r *resourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "The resource `random_pet` generates random pet names that are intended to be used as " +
			"unique identifiers for other resources.\n" +
			"\n" +
			"This resource can be used in conjunction with resources that have the `create_before_destroy` " +
			"lifecycle flag set, to avoid conflicts with unique names during the brief period where both the old " +
			"and new resources exist concurrently.",
		Attributes: map[string]tfsdk.Attribute{
			"keepers": {
				Description: "Arbitrary map of values that, when changed, will trigger recreation of " +
					"resource. See [the main provider documentation](../index.html) for more information.",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"length": {
				Description: "The length (in words) of the pet name. Defaults to 2",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.Int64{Value: 2}),
					planmodifiers.RequiresReplace(),
				},
			},
			"prefix": {
				Description:   "A string to prefix the name with.",
				Type:          types.StringType,
				Optional:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"separator": {
				Description: "The character to separate words in the pet name. Defaults to \"-\"",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.DefaultValue(types.String{Value: "-"}),
					planmodifiers.RequiresReplace(),
				},
			},
			"id": {
				Description: "The random pet name.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (r *resourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return &resource{}, nil
}

var _ tfsdk.Resource = (*resource)(nil)

type resource struct{}

func (r *resource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// This is necessary to ensure each call to petname is properly randomised:
	// the library uses `rand.Intn()` and does NOT seed `rand.Seed()` by default,
	// so this call takes care of that.
	petname.NonDeterministicMode()

	var plan modelV0

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	length := plan.Length.Value
	separator := plan.Separator.Value
	prefix := plan.Prefix.Value

	pet := strings.ToLower(petname.Generate(int(length), separator))

	pn := modelV0{
		Keepers:   plan.Keepers,
		Length:    types.Int64{Value: length},
		Separator: types.String{Value: separator},
	}

	if prefix != "" {
		pet = fmt.Sprintf("%s%s%s", prefix, separator, pet)
		pn.Prefix.Value = prefix
	} else {
		pn.Prefix.Null = true
	}

	pn.ID.Value = pet

	diags = resp.State.Set(ctx, pn)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *resource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}

// Update is intentionally left blank as all required and optional attributes force replacement of the resource
// through the RequiresReplace AttributePlanModifier.
func (r *resource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *resource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

type modelV0 struct {
	ID        types.String `tfsdk:"id"`
	Keepers   types.Map    `tfsdk:"keepers"`
	Length    types.Int64  `tfsdk:"length"`
	Prefix    types.String `tfsdk:"prefix"`
	Separator types.String `tfsdk:"separator"`
}
