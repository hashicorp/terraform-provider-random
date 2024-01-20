// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-random/internal/diagnostics"
	mapplanmodifiers "github.com/terraform-providers/terraform-provider-random/internal/planmodifiers/map"
)

var (
	_ resource.Resource = (*ipResource)(nil)
)

func NewIPResource() resource.Resource {
	return &ipResource{}
}

type ipResource struct{}

func (r *ipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "ip"
}

func (r *ipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `random_ip` generates a random IP address from a given CIDR range based on the " +
			"address type specified.",
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
			"address_type": schema.StringAttribute{
				Description: "A string indicating the type of IP address to generate. Valid values are `ipv4` and `ipv6`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"ipv4", "ipv6"}...),
				},
			},
			"cidr_range": schema.StringAttribute{
				Description: "A CIDR range from which to allocate the IP address.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"result": schema.StringAttribute{
				Description: "The random IP address allocated from the CIDR range.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *ipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipModelV0

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	addressType := plan.AddressType.ValueString()
	cidrRange := plan.CIDRRange.ValueString()

	if valid, err := isValidCIDRRange(cidrRange); !valid {
		resp.Diagnostics.AddError(
			"Create Random IP error",
			"There was an error when validating the CIDR range.\n\n"+
				diagnostics.RetryMsg+
				"Original Error: "+err.Error(),
		)
		return
	}

	u := &ipModelV0{
		ID:          types.StringValue("-"),
		Keepers:     plan.Keepers,
		AddressType: types.StringValue(addressType),
		CIDRRange:   types.StringValue(cidrRange),
	}

	diags = resp.State.Set(ctx, u)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *ipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update ensures the plan value is copied to the state to complete the update.
func (r *ipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model ipModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *ipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

type ipModelV0 struct {
	ID          types.String `tfsdk:"id"`
	Keepers     types.Map    `tfsdk:"keepers"`
	AddressType types.String `tfsdk:"address_type"`
	CIDRRange   types.String `tfsdk:"cidr_range"`
	Result      types.String `tfsdk:"result"`
}

func isValidCIDRRange(cidrRange string) (bool, error) {
	_, _, err := net.ParseCIDR(cidrRange)
	return err == nil, err
}
