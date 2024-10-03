package provider

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"golang.org/x/crypto/bcrypt"
)

// This test creates a low-level protocol request to OpenEphemeralResource and then asserts
// that the generated password matches the expected length and the bcrypt_hash is valid.
func TestAccEphemeralResourcePassword_Result(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Get the random provider instance
	providerFunc := providerserver.NewProtocol5WithError(New())
	randomProvider, err := providerFunc()
	if err != nil {
		t.Fatalf("error retrieving random provider: %s", err)
	}

	configPasswordLength := 12
	configObj := map[string]tftypes.Value{
		"length":           tftypes.NewValue(tftypes.Number, configPasswordLength),
		"special":          tftypes.NewValue(tftypes.Bool, nil),
		"upper":            tftypes.NewValue(tftypes.Bool, nil),
		"lower":            tftypes.NewValue(tftypes.Bool, nil),
		"numeric":          tftypes.NewValue(tftypes.Bool, nil),
		"min_numeric":      tftypes.NewValue(tftypes.Number, nil),
		"min_upper":        tftypes.NewValue(tftypes.Number, nil),
		"min_lower":        tftypes.NewValue(tftypes.Number, nil),
		"min_special":      tftypes.NewValue(tftypes.Number, nil),
		"override_special": tftypes.NewValue(tftypes.String, nil),
		"result":           tftypes.NewValue(tftypes.String, nil),
		"bcrypt_hash":      tftypes.NewValue(tftypes.String, nil),
	}

	// Get the random_password ephemeral resource schema
	ephPasswordSchema := &passwordEphemeralResource{}
	schemaResp := ephemeral.SchemaResponse{}
	ephPasswordSchema.Schema(ctx, ephemeral.SchemaRequest{}, &schemaResp)
	ephemeralPasswordSchemaType := schemaResp.Schema.Type().TerraformType(ctx)

	// TODO: Provider server implementation is optional for the first release of ephemeral resources to avoid build errors during dependency updates
	// nolint: staticcheck
	ephSupportedProvider := randomProvider.(tfprotov5.ProviderServerWithEphemeralResources)
	req := &tfprotov5.OpenEphemeralResourceRequest{
		TypeName: "random_password",
		Config: testNewDynamicValueMust(
			t,
			ephemeralPasswordSchemaType,
			tftypes.NewValue(ephemeralPasswordSchemaType, configObj),
		),
	}

	// Execute the RPC call (this is about as close as we can get to testing fully what has been implemented so far)
	gotResp, err := ephSupportedProvider.OpenEphemeralResource(ctx, req)
	if err != nil {
		t.Fatalf("error executing OpenEphemeralResource: %s", err)
	}

	stateValue, err := gotResp.Result.Unmarshal(ephemeralPasswordSchemaType)
	if err != nil {
		t.Fatalf("error parsing MsgPack response from OpenEphemeralResource: %s", err)
	}

	// Validate the result attribute length matches expectation
	resultValue, _ := stateValue.ApplyTerraform5AttributePathStep(tftypes.AttributeName("result"))
	resultTfValue := resultValue.(tftypes.Value) // nolint

	var resultStr string
	resultTfValue.As(&resultStr) // nolint

	if configPasswordLength != len(resultStr) {
		t.Fatalf("expected 'result' to be length of: %d, got: %d", configPasswordLength, len(resultStr))
	}

	// Validate the bcrypt_hash attribute is a valid match of the result attribute
	bcryptHashValue, _ := stateValue.ApplyTerraform5AttributePathStep(tftypes.AttributeName("bcrypt_hash"))
	bcryptHashTfValue := bcryptHashValue.(tftypes.Value) // nolint

	var bcryptHashStr string
	bcryptHashTfValue.As(&bcryptHashStr) // nolint

	if err := bcrypt.CompareHashAndPassword([]byte(bcryptHashStr), []byte(resultStr)); err != nil {
		t.Fatalf("bcrypt_hash %q is not a valid hash of the result %q: %s", bcryptHashStr, resultStr, err)
	}
}

// This test creates a low-level protocol request to ValidateEphemeralResourceConfig for each test case
// and then asserts that the expected diagnostics match.
func TestAccEphemeralResourcePassword_Validation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Get the random provider instance
	providerFunc := providerserver.NewProtocol5WithError(New())
	randomProvider, err := providerFunc()
	if err != nil {
		t.Fatalf("error retrieving random provider: %s", err)
	}

	// Get the random_password ephemeral resource schema
	ephPasswordSchema := &passwordEphemeralResource{}
	schemaResp := ephemeral.SchemaResponse{}
	ephPasswordSchema.Schema(ctx, ephemeral.SchemaRequest{}, &schemaResp)
	ephemeralPasswordSchemaType := schemaResp.Schema.Type().TerraformType(ctx)

	testCases := map[string]struct {
		config        map[string]tftypes.Value
		expectedDiags []*tfprotov5.Diagnostic
	}{
		"valid_config": {
			config: map[string]tftypes.Value{
				"length":           tftypes.NewValue(tftypes.Number, 12),
				"special":          tftypes.NewValue(tftypes.Bool, true),
				"upper":            tftypes.NewValue(tftypes.Bool, true),
				"lower":            tftypes.NewValue(tftypes.Bool, true),
				"numeric":          tftypes.NewValue(tftypes.Bool, true),
				"min_numeric":      tftypes.NewValue(tftypes.Number, 1),
				"min_upper":        tftypes.NewValue(tftypes.Number, 2),
				"min_lower":        tftypes.NewValue(tftypes.Number, 3),
				"min_special":      tftypes.NewValue(tftypes.Number, 4),
				"override_special": tftypes.NewValue(tftypes.String, nil),
				"result":           tftypes.NewValue(tftypes.String, nil),
				"bcrypt_hash":      tftypes.NewValue(tftypes.String, nil),
			},
		},
		"valid_config_override": {
			config: map[string]tftypes.Value{
				"length":           tftypes.NewValue(tftypes.Number, 4),
				"special":          tftypes.NewValue(tftypes.Bool, nil),
				"upper":            tftypes.NewValue(tftypes.Bool, false),
				"lower":            tftypes.NewValue(tftypes.Bool, false),
				"numeric":          tftypes.NewValue(tftypes.Bool, false),
				"min_numeric":      tftypes.NewValue(tftypes.Number, nil),
				"min_upper":        tftypes.NewValue(tftypes.Number, nil),
				"min_lower":        tftypes.NewValue(tftypes.Number, nil),
				"min_special":      tftypes.NewValue(tftypes.Number, nil),
				"override_special": tftypes.NewValue(tftypes.String, "!"),
				"result":           tftypes.NewValue(tftypes.String, nil),
				"bcrypt_hash":      tftypes.NewValue(tftypes.String, nil),
			},
		},
		"invalid_length": {
			config: map[string]tftypes.Value{
				"length":           tftypes.NewValue(tftypes.Number, 0),
				"special":          tftypes.NewValue(tftypes.Bool, nil),
				"upper":            tftypes.NewValue(tftypes.Bool, nil),
				"lower":            tftypes.NewValue(tftypes.Bool, nil),
				"numeric":          tftypes.NewValue(tftypes.Bool, nil),
				"min_numeric":      tftypes.NewValue(tftypes.Number, nil),
				"min_upper":        tftypes.NewValue(tftypes.Number, nil),
				"min_lower":        tftypes.NewValue(tftypes.Number, nil),
				"min_special":      tftypes.NewValue(tftypes.Number, nil),
				"override_special": tftypes.NewValue(tftypes.String, nil),
				"result":           tftypes.NewValue(tftypes.String, nil),
				"bcrypt_hash":      tftypes.NewValue(tftypes.String, nil),
			},
			expectedDiags: []*tfprotov5.Diagnostic{
				{
					Severity:  tfprotov5.DiagnosticSeverityError,
					Summary:   "Invalid Attribute Value",
					Detail:    "Attribute length value must be at least 1, got: 0",
					Attribute: tftypes.NewAttributePath().WithAttributeName("length"),
				},
			},
		},
		"invalid_length_sum": {
			config: map[string]tftypes.Value{
				"length":           tftypes.NewValue(tftypes.Number, 9),
				"special":          tftypes.NewValue(tftypes.Bool, nil),
				"upper":            tftypes.NewValue(tftypes.Bool, nil),
				"lower":            tftypes.NewValue(tftypes.Bool, nil),
				"numeric":          tftypes.NewValue(tftypes.Bool, nil),
				"min_numeric":      tftypes.NewValue(tftypes.Number, 1),
				"min_upper":        tftypes.NewValue(tftypes.Number, 2),
				"min_lower":        tftypes.NewValue(tftypes.Number, 3),
				"min_special":      tftypes.NewValue(tftypes.Number, 4),
				"override_special": tftypes.NewValue(tftypes.String, nil),
				"result":           tftypes.NewValue(tftypes.String, nil),
				"bcrypt_hash":      tftypes.NewValue(tftypes.String, nil),
			},
			expectedDiags: []*tfprotov5.Diagnostic{
				{
					Severity:  tfprotov5.DiagnosticSeverityError,
					Summary:   "Invalid Attribute Value",
					Detail:    "Attribute length value must be at least sum of min_upper + min_lower + min_numeric + min_special, got: 9",
					Attribute: tftypes.NewAttributePath().WithAttributeName("length"),
				},
			},
		},
		"invalid_constraint_combination": {
			config: map[string]tftypes.Value{
				"length":           tftypes.NewValue(tftypes.Number, 12),
				"special":          tftypes.NewValue(tftypes.Bool, false),
				"upper":            tftypes.NewValue(tftypes.Bool, false),
				"lower":            tftypes.NewValue(tftypes.Bool, false),
				"numeric":          tftypes.NewValue(tftypes.Bool, false),
				"min_numeric":      tftypes.NewValue(tftypes.Number, nil),
				"min_upper":        tftypes.NewValue(tftypes.Number, nil),
				"min_lower":        tftypes.NewValue(tftypes.Number, nil),
				"min_special":      tftypes.NewValue(tftypes.Number, nil),
				"override_special": tftypes.NewValue(tftypes.String, nil),
				"result":           tftypes.NewValue(tftypes.String, nil),
				"bcrypt_hash":      tftypes.NewValue(tftypes.String, nil),
			},
			expectedDiags: []*tfprotov5.Diagnostic{
				{
					Severity:  tfprotov5.DiagnosticSeverityError,
					Summary:   "Invalid Attribute Combination",
					Detail:    "At least one attribute out of [special,upper,lower,numeric] must be specified as true",
					Attribute: tftypes.NewAttributePath().WithAttributeName("numeric"),
				},
			},
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// TODO: Provider server implementation is optional for the first release of ephemeral resources to avoid build errors during dependency updates
			// nolint: staticcheck
			ephSupportedProvider := randomProvider.(tfprotov5.ProviderServerWithEphemeralResources)
			req := &tfprotov5.ValidateEphemeralResourceConfigRequest{
				TypeName: "random_password",
				Config: testNewDynamicValueMust(
					t,
					ephemeralPasswordSchemaType,
					tftypes.NewValue(ephemeralPasswordSchemaType, testCase.config),
				),
			}

			// Execute the RPC call (this is about as close as we can get to testing fully what has been implemented so far)
			gotResp, err := ephSupportedProvider.ValidateEphemeralResourceConfig(ctx, req)
			if err != nil {
				t.Fatalf("error executing ValidateEphemeralResourceConfig: %s", err)
			}

			if diff := cmp.Diff(gotResp.Diagnostics, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func testNewDynamicValueMust(t *testing.T, typ tftypes.Type, value tftypes.Value) *tfprotov5.DynamicValue {
	t.Helper()

	dynamicValue, err := tfprotov5.NewDynamicValue(typ, value)

	if err != nil {
		t.Fatalf("unable to create DynamicValue: %s", err)
	}

	return &dynamicValue
}
