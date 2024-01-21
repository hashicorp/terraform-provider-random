// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
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
	resp.TypeName = req.ProviderTypeName + "_ip"
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
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"ipv4", "ipv6"}...),
				},
			},
			"cidr_range": schema.StringAttribute{
				Description: "A CIDR range from which to allocate the IP address.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

	// Generate a random IP address from a given CIDR range.
	ip, err := getRandomIP(cidrRange, addressType)
	if err != nil {
		resp.Diagnostics.AddError(
			"Create Random IP error",
			"There was an error when generating the random IP address.\n\n"+
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
		Result:      types.StringValue(ip.String()),
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

func getRandomIP(cidrRange, addressType string) (net.IP, error) {
	// We first parse the CIDR range to check for errors
	// and to get the network address and netmask.
	_, network, err := net.ParseCIDR(cidrRange)
	if err != nil {
		return nil, fmt.Errorf("error parsing CIDR range: %w", err)
	}

	// We then check the length of the netmask to see if it is valid.
	netmask := network.Mask
	if len(netmask) != net.IPv4len && len(netmask) != net.IPv6len {
		return nil, fmt.Errorf("invalid netmask length: %d", len(netmask))
	}

	// Get the network address as a byte slice, either in 4 or 16 byte form.
	var address net.IP
	switch addressType {
	case "ipv4":
		address = network.IP.To4()
	case "ipv6":
		address = network.IP.To16()
	default:
		return nil, fmt.Errorf("invalid address type: %s", addressType)
	}

	var picked []byte
	for i := 0; i < len(netmask); i++ {
		// Combine the random bits, determined by the XOR of 255 with the netmask byte
		// and a randomly generated byte, with the original bits in the network address.
		// The bitwise OR operation ensures that bits are set where the netmask has 0s,
		// introducing randomness, while retaining the original bits where the netmask has 1s.
		picked = append(picked, ((255^netmask[i])&byte(rand.Intn(256)))|address[i])
	}

	// Turn the randomized byte slice into an IP address.
	pickedIP := net.IP(picked)

	return pickedIP, nil
}
