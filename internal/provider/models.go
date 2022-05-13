package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type IDModel struct {
	ID         types.String `tfsdk:"id"`
	Keepers    types.Map    `tfsdk:"keepers"`
	ByteLength types.Int64  `tfsdk:"byte_length"`
	Prefix     types.String `tfsdk:"prefix"`
	B64URL     types.String `tfsdk:"b64_url"`
	B64Std     types.String `tfsdk:"b64_std"`
	Hex        types.String `tfsdk:"hex"`
	Dec        types.String `tfsdk:"dec"`
}

type IntegerModel struct {
	ID      types.String `tfsdk:"id"`
	Keepers types.Map    `tfsdk:"keepers"`
	Min     types.Int64  `tfsdk:"min"`
	Max     types.Int64  `tfsdk:"max"`
	Seed    types.String `tfsdk:"seed"`
	Result  types.Int64  `tfsdk:"result"`
}

type PasswordModel struct {
	ID              types.String `tfsdk:"id"`
	Keepers         types.Map    `tfsdk:"keepers"`
	Length          types.Int64  `tfsdk:"length"`
	Special         types.Bool   `tfsdk:"special"`
	Upper           types.Bool   `tfsdk:"upper"`
	Lower           types.Bool   `tfsdk:"lower"`
	Number          types.Bool   `tfsdk:"number"`
	MinNumeric      types.Int64  `tfsdk:"min_numeric"`
	MinUpper        types.Int64  `tfsdk:"min_upper"`
	MinLower        types.Int64  `tfsdk:"min_lower"`
	MinSpecial      types.Int64  `tfsdk:"min_special"`
	OverrideSpecial types.String `tfsdk:"override_special"`
	Result          types.String `tfsdk:"result"`
	BcryptHash      types.String `tfsdk:"bcrypt_hash"`
}

func (pm PasswordModel) LengthValue() int64 {
	return pm.Length.Value
}

type PetNameModel struct {
	ID        types.String `tfsdk:"id"`
	Keepers   types.Map    `tfsdk:"keepers"`
	Length    types.Int64  `tfsdk:"length"`
	Prefix    types.String `tfsdk:"prefix"`
	Separator types.String `tfsdk:"separator"`
}

type ShuffleModel struct {
	ID          types.String `tfsdk:"id"`
	Keepers     types.Map    `tfsdk:"keepers"`
	Seed        types.String `tfsdk:"seed"`
	Input       types.List   `tfsdk:"input"`
	ResultCount types.Int64  `tfsdk:"result_count"`
	Result      types.List   `tfsdk:"result"`
}

type StringModel struct {
	ID              types.String `tfsdk:"id"`
	Keepers         types.Map    `tfsdk:"keepers"`
	Length          types.Int64  `tfsdk:"length"`
	Special         types.Bool   `tfsdk:"special"`
	Upper           types.Bool   `tfsdk:"upper"`
	Lower           types.Bool   `tfsdk:"lower"`
	Number          types.Bool   `tfsdk:"number"`
	MinNumeric      types.Int64  `tfsdk:"min_numeric"`
	MinUpper        types.Int64  `tfsdk:"min_upper"`
	MinLower        types.Int64  `tfsdk:"min_lower"`
	MinSpecial      types.Int64  `tfsdk:"min_special"`
	OverrideSpecial types.String `tfsdk:"override_special"`
	Result          types.String `tfsdk:"result"`
}

type UUIDModel struct {
	ID      types.String `tfsdk:"id"`
	Keepers types.Map    `tfsdk:"keepers"`
	Result  types.String `tfsdk:"result"`
}
