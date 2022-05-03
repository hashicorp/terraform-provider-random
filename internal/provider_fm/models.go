package provider_fm

import "github.com/hashicorp/terraform-plugin-framework/types"

type ID struct {
	ID         types.String `tfsdk:"id"`
	Keepers    types.Map    `tfsdk:"keepers"`
	ByteLength types.Int64  `tfsdk:"byte_length"`
	Prefix     types.String `tfsdk:"prefix"`
	B64URL     types.String `tfsdk:"b64_url"`
	B64Std     types.String `tfsdk:"b64_std"`
	Hex        types.String `tfsdk:"hex"`
	Dec        types.String `tfsdk:"dec"`
}

type UUID struct {
	ID      types.String `tfsdk:"id"`
	Result  types.String `tfsdk:"result"`
	Keepers types.Map    `tfsdk:"keepers"`
}
