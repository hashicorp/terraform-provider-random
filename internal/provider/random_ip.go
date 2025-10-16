// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
		Description: "The `random_ip` resource generates a random IP address, either IPv4 or IPv6. By default, it randomly chooses between 0.0.0.0/0 (IPv4) and ::/0 (IPv6). You can influence the IP type by specifying a `cidr_range`.",
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
			"cidr_range": schema.StringAttribute{
				Description: "A CIDR range from which to allocate the IP address.",
				Computed:    true,
				Optional:    true,
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

	cidrRange := plan.CIDRRange.ValueString()
	// Check if CIDR range is empty, and if so, set it to either 0.0.0.0/0 or ::/0
	if cidrRange == "" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		if r.Intn(2) == 0 {
			cidrRange = "0.0.0.0/0"
		} else {
			cidrRange = "::/0"
		}
	}

	// Generate a random IP address from a given CIDR range.
	ip, err := getRandomIP(cidrRange)
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
		ID:        types.StringValue("-"),
		Keepers:   plan.Keepers,
		CIDRRange: types.StringValue(cidrRange),
		Result:    types.StringValue(ip.String()),
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
	ID        types.String `tfsdk:"id"`
	Keepers   types.Map    `tfsdk:"keepers"`
	CIDRRange types.String `tfsdk:"cidr_range"`
	Result    types.String `tfsdk:"result"`
}

func getRandomIP(cidrRange string) (net.IP, error) {
	// Parse the CIDR range to check for errors
	// and to get the network address and netmask.
	_, network, err := net.ParseCIDR(cidrRange)
	if err != nil {
		return nil, fmt.Errorf("error parsing CIDR range: %w", err)
	}

	// Get the netmask and determine the IP type (IPv4 or IPv6)
	netmaskBytes := network.Mask
	addressBytes := network.IP

	// Check if the length of the netmask is valid (either IPv4 or IPv6 length).
	if len(netmaskBytes) != net.IPv4len && len(netmaskBytes) != net.IPv6len {
		return nil, fmt.Errorf("invalid netmask: must be either IPv4 or IPv6")
	}

	// This typically occurs when the CIDR range is not of the same type as the address type.
	if len(netmaskBytes) != len(addressBytes) {
		return nil, fmt.Errorf("netmask byte length does not match IP address byte length")
	}

	// Generate the random IP within the CIDR range.
	picked := make([]byte, len(addressBytes))
	for i := 0; i < len(netmaskBytes); i++ {
		// Combine the random bits with the original network address
		picked[i] = ((255 ^ netmaskBytes[i]) & byte(rand.Intn(256))) | addressBytes[i]
	}

	// Turn the randomized byte slice into an IP address.
	pickedIP := net.IP(picked)

	return pickedIP, nil
}
