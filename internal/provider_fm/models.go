package provider_fm

import "github.com/hashicorp/terraform-plugin-framework/types"

type UUID struct {
	ID     types.String `tfsdk:"id"`
	Result types.String `tfsdk:"result"`
}
