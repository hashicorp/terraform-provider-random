package provider

import (
	"context"
	"fmt"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type resourcePetType struct{}

func (r resourcePetType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	// This is necessary to ensure each call to petname is properly randomised:
	// the library uses `rand.Intn()` and does NOT seed `rand.Seed()` by default,
	// so this call takes care of that.
	petname.NonDeterministicMode()

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
				Optional:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"length": {
				Description: "The length (in words) of the pet name. Defaults to 2",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
					defaultInt(2),
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
					tfsdk.RequiresReplace(),
					defaultString("-"),
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

func (r resourcePetType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourcePet{
		p: *(p.(*provider)),
	}, nil
}

type resourcePet struct {
	p provider
}

func (r resourcePet) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan PetNameModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	length := plan.Length.Value
	separator := plan.Separator.Value
	prefix := plan.Prefix.Value

	pet := strings.ToLower(petname.Generate(int(length), separator))

	pn := PetNameModel{
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

func (r resourcePet) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Intentionally left blank.
}

func (r resourcePet) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Intentionally left blank.
}

func (r resourcePet) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	resp.State.RemoveResource(ctx)
}
