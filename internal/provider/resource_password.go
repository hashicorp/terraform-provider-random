package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/crypto/bcrypt"
)

func resourcePassword() *schema.Resource {
	return &schema.Resource{
		Description: "Identical to [random_string](string.html) with the exception that the result is " +
			"treated as sensitive and, thus, _not_ displayed in console output. Read more about sensitive " +
			"data handling in the [Terraform documentation](https://www.terraform.io/docs/language/state/sensitive-data.html).\n" +
			"\n" +
			"This resource *does* use a cryptographic random number generator.",
		CreateContext: createPassword,
		ReadContext:   readNil,
		DeleteContext: RemoveResourceFromState,
		Schema:        passwordSchemaV2(),
		Importer: &schema.ResourceImporter{
			StateContext: importPasswordFunc,
		},
		SchemaVersion: 2,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    resourcePasswordV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourcePasswordStateUpgradeV0,
			},
			{
				Version: 1,
				Type:    resourcePasswordV1().CoreConfigSchema().ImpliedType(),
				Upgrade: resourcePasswordStateUpgradeV1,
			},
		},
	}
}

func createPassword(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	diags := createStringFunc(true)(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	hash, err := generateHash(d.Get("result").(string))
	if err != nil {
		diags = append(diags, diag.Errorf("err: %s", err)...)
		return diags
	}

	if err := d.Set("bcrypt_hash", hash); err != nil {
		diags = append(diags, diag.Errorf("err: %s", err)...)
		return diags
	}

	return nil
}

func importPasswordFunc(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	val := d.Id()
	d.SetId("none")

	if err := d.Set("result", val); err != nil {
		return nil, fmt.Errorf("resource password import failed, error setting result: %w", err)
	}

	hash, err := generateHash(val)
	if err != nil {
		return nil, fmt.Errorf("resource password import failed, generate hash error: %w", err)
	}

	if err := d.Set("bcrypt_hash", hash); err != nil {
		return nil, fmt.Errorf("resource password import failed, error setting bcrypt_hash: %w", err)
	}

	return []*schema.ResourceData{d}, nil
}

func resourcePasswordV1() *schema.Resource {
	return &schema.Resource{
		Schema: passwordSchemaV1(),
	}
}

func resourcePasswordStateUpgradeV1(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		return nil, fmt.Errorf("resource password state upgrade failed, state is nil")
	}

	number, ok := rawState["number"].(bool)
	if !ok {
		return nil, fmt.Errorf("resource password state upgrade failed, number could not be asserted as bool: %T", rawState["number"])
	}

	rawState["numeric"] = number

	return rawState, nil
}

func resourcePasswordV0() *schema.Resource {
	return &schema.Resource{
		Schema: passwordSchemaV0(),
	}
}

func resourcePasswordStateUpgradeV0(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		return nil, fmt.Errorf("resource password state upgrade failed, state is nil")
	}

	result, ok := rawState["result"].(string)
	if !ok {
		return nil, fmt.Errorf("resource password state upgrade failed, result could not be asserted as string: %T", rawState["result"])
	}

	hash, err := generateHash(result)
	if err != nil {
		return nil, fmt.Errorf("resource password state upgrade failed, generate hash error: %w", err)
	}

	rawState["bcrypt_hash"] = hash

	return rawState, nil
}

func generateHash(toHash string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(toHash), bcrypt.DefaultCost)

	return string(hash), err
}
