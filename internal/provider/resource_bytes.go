package provider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-random/internal/diagnostics"
)

var (
	_ resource.Resource                = (*bytesResource)(nil)
	_ resource.ResourceWithImportState = (*bytesResource)(nil)
)

func NewBytesResource() resource.Resource {
	return &bytesResource{}
}

type bytesResource struct {
}

func (r *bytesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bytes"
}

func (r *bytesResource) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return bytesSchemaV0(), nil
}

func (r *bytesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bytesModelV0

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bytes := make([]byte, plan.Length.ValueInt64())
	_, err := rand.Read(bytes)
	if err != nil {
		resp.Diagnostics.AddError(
			"Create Random bytes error",
			"There was an error during random generation.\n\n"+
				diagnostics.RetryMsg+
				fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	u := &bytesModelV0{
		ID:           types.StringValue("none"),
		Length:       plan.Length,
		ResultBase64: types.StringValue(base64.StdEncoding.EncodeToString(bytes)),
		ResultHex:    types.StringValue(hex.EncodeToString(bytes)),
		Keepers:      plan.Keepers,
	}

	diags = resp.State.Set(ctx, u)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *bytesResource) Read(context.Context, resource.ReadRequest, *resource.ReadResponse) {
}

// Update is intentionally left blank as all required and optional attributes force replacement of the resource
// through the RequiresReplace AttributePlanModifier.
func (r *bytesResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *bytesResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *bytesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	bytes, err := base64.StdEncoding.DecodeString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Import Random bytes Error",
			"There was an error during the parsing of the base64 string.\n\n"+
				diagnostics.RetryMsg+
				fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	var state bytesModelV0

	state.ID = types.StringValue("none")
	state.Length = types.Int64Value(int64(len(bytes)))
	state.ResultBase64 = types.StringValue(req.ID)
	state.ResultHex = types.StringValue(hex.EncodeToString(bytes))
	state.Keepers = types.MapValueMust(types.StringType, nil)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

type bytesModelV0 struct {
	ID           types.String `tfsdk:"id"`
	Length       types.Int64  `tfsdk:"length"`
	Keepers      types.Map    `tfsdk:"keepers"`
	ResultBase64 types.String `tfsdk:"result_base64"`
	ResultHex    types.String `tfsdk:"result_hex"`
}

func bytesSchemaV0() tfsdk.Schema {
	return tfsdk.Schema{
		Version: 0,
		Description: "The resource `random_bytes` generates random bytes that are intended to be " +
			"used as secret or keys.",
		Attributes: map[string]tfsdk.Attribute{
			"keepers": {
				Description: "Arbitrary map of values that, when changed, will trigger recreation of " +
					"resource. See [the main provider documentation](../index.html) for more information.",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"length": {
				Description: "The number of bytes requested. The minimum value for length is 1.",
				Type:        types.Int64Type,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
				Validators: []tfsdk.AttributeValidator{
					int64validator.AtLeast(1),
				},
			},
			"result_base64": {
				Description: "The generated bytes presented in base64 string format.",
				Type:        types.StringType,
				Computed:    true,
				Sensitive:   true,
			},
			"result_hex": {
				Description: "The generated bytes presented in hex string format.",
				Type:        types.StringType,
				Computed:    true,
				Sensitive:   true,
			},
			"id": {
				Description: "A static value used internally by Terraform, this should not be referenced in configurations.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}
}
