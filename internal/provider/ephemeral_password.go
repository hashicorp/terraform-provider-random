// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-random/internal/diagnostics"
	"github.com/terraform-providers/terraform-provider-random/internal/random"
	"github.com/terraform-providers/terraform-provider-random/internal/validators"
)

var (
	_ ephemeral.EphemeralResource = (*passwordEphemeralResource)(nil)
)

func NewPasswordEphemeralResource() ephemeral.EphemeralResource {
	return &passwordEphemeralResource{}
}

type passwordEphemeralResource struct{}

type ephemeralPasswordModel struct {
	Length          types.Int64  `tfsdk:"length"`
	Special         types.Bool   `tfsdk:"special"`
	Upper           types.Bool   `tfsdk:"upper"`
	Lower           types.Bool   `tfsdk:"lower"`
	Numeric         types.Bool   `tfsdk:"numeric"`
	MinNumeric      types.Int64  `tfsdk:"min_numeric"`
	MinUpper        types.Int64  `tfsdk:"min_upper"`
	MinLower        types.Int64  `tfsdk:"min_lower"`
	MinSpecial      types.Int64  `tfsdk:"min_special"`
	OverrideSpecial types.String `tfsdk:"override_special"`
	Result          types.String `tfsdk:"result"`
	BcryptHash      types.String `tfsdk:"bcrypt_hash"`
}

func (e *passwordEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password"
}

func (e *passwordEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "-> If the managed resource doesn't have a write-only attribute available for the password (first introduced in Terraform 1.11), then the " +
			"password can only be created with the managed resource variant of [`random_password`](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/password).\n" +
			"\n" +
			"Generates an ephemeral password string using a cryptographic random number generator.\n" +
			"\n" +
			"The primary use-case for generating an ephemeral random password is to be used in combination with a write-only attribute " +
			"in a managed resource, which will avoid Terraform storing the password string in the plan or state file.",
		Attributes: map[string]schema.Attribute{
			"length": schema.Int64Attribute{
				Description: "The length of the string desired. The minimum value for length is 1 and, length " +
					"must also be >= (`min_upper` + `min_lower` + `min_numeric` + `min_special`).",
				Required: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
					int64validator.AtLeastSumOf(
						path.MatchRoot("min_upper"),
						path.MatchRoot("min_lower"),
						path.MatchRoot("min_numeric"),
						path.MatchRoot("min_special"),
					),
				},
			},
			"special": schema.BoolAttribute{
				Description: "Include special characters in the result. These are `!@#$%&*()-_=+[]{}<>:?`. Default value is `true`.",
				Optional:    true,
				Computed:    true,
			},
			"upper": schema.BoolAttribute{
				Description: "Include uppercase alphabet characters in the result. Default value is `true`.",
				Optional:    true,
				Computed:    true,
			},
			"lower": schema.BoolAttribute{
				Description: "Include lowercase alphabet characters in the result. Default value is `true`.",
				Optional:    true,
				Computed:    true,
			},
			"numeric": schema.BoolAttribute{
				Description: "Include numeric characters in the result. Default value is `true`. " +
					"If `numeric`, `upper`, `lower`, and `special` are all configured, at least one " +
					"of them must be set to `true`.",
				Optional: true,
				Computed: true,
				Validators: []validator.Bool{
					validators.AtLeastOneOfTrue(
						path.MatchRoot("special"),
						path.MatchRoot("upper"),
						path.MatchRoot("lower"),
					),
				},
			},
			"min_numeric": schema.Int64Attribute{
				Description: "Minimum number of numeric characters in the result. Default value is `0`.",
				Optional:    true,
				Computed:    true,
			},

			"min_upper": schema.Int64Attribute{
				Description: "Minimum number of uppercase alphabet characters in the result. Default value is `0`.",
				Optional:    true,
				Computed:    true,
			},
			"min_lower": schema.Int64Attribute{
				Description: "Minimum number of lowercase alphabet characters in the result. Default value is `0`.",
				Optional:    true,
				Computed:    true,
			},
			"min_special": schema.Int64Attribute{
				Description: "Minimum number of special characters in the result. Default value is `0`.",
				Optional:    true,
				Computed:    true,
			},
			"override_special": schema.StringAttribute{
				Description: "Supply your own list of special characters to use for string generation.  This " +
					"overrides the default character list in the special argument.  The `special` argument must " +
					"still be set to true for any overwritten characters to be used in generation.",
				Optional: true,
			},
			"result": schema.StringAttribute{
				Description: "The generated random string.",
				Computed:    true,
				Sensitive:   true,
			},
			"bcrypt_hash": schema.StringAttribute{
				Description: "A bcrypt hash of the generated random string. " +
					"**NOTE**: If the generated random string is greater than 72 bytes in length, " +
					"`bcrypt_hash` will contain a hash of the first 72 bytes.",
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func (e *passwordEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data ephemeralPasswordModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	applyDefaultPasswordParameters(&data)

	params := random.StringParams{
		Length:          data.Length.ValueInt64(),
		Upper:           data.Upper.ValueBool(),
		MinUpper:        data.MinUpper.ValueInt64(),
		Lower:           data.Lower.ValueBool(),
		MinLower:        data.MinLower.ValueInt64(),
		Numeric:         data.Numeric.ValueBool(),
		MinNumeric:      data.MinNumeric.ValueInt64(),
		Special:         data.Special.ValueBool(),
		MinSpecial:      data.MinSpecial.ValueInt64(),
		OverrideSpecial: data.OverrideSpecial.ValueString(),
	}

	result, err := random.CreateString(params)
	if err != nil {
		resp.Diagnostics.Append(diagnostics.RandomReadError(err.Error())...)
		return
	}

	hash, err := generateHash(string(result))
	if err != nil {
		resp.Diagnostics.Append(diagnostics.HashGenerationError(err.Error())...)
	}

	data.BcryptHash = types.StringValue(hash)
	data.Result = types.StringValue(string(result))

	resp.Diagnostics.Append(resp.Result.Set(ctx, data)...)
}

func applyDefaultPasswordParameters(data *ephemeralPasswordModel) {
	if data.Special.IsNull() {
		data.Special = types.BoolValue(true)
	}
	if data.Upper.IsNull() {
		data.Upper = types.BoolValue(true)
	}
	if data.Lower.IsNull() {
		data.Lower = types.BoolValue(true)
	}
	if data.Numeric.IsNull() {
		data.Numeric = types.BoolValue(true)
	}

	if data.MinNumeric.IsNull() {
		data.MinNumeric = types.Int64Value(0)
	}
	if data.MinUpper.IsNull() {
		data.MinUpper = types.Int64Value(0)
	}
	if data.MinLower.IsNull() {
		data.MinLower = types.Int64Value(0)
	}
	if data.MinSpecial.IsNull() {
		data.MinSpecial = types.Int64Value(0)
	}
}
