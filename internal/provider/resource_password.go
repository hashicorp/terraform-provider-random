package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/crypto/bcrypt"
)

func resourcePassword() *schema.Resource {
	passwordSchema := stringSchemaV1(true)
	passwordSchema["bcrypt_hash"] = &schema.Schema{
		Description: "A bcrypt hash of the generated random string.",
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
	}

	create := func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	return &schema.Resource{
		Description: "Identical to [random_string](string.html) with the exception that the result is " +
			"treated as sensitive and, thus, _not_ displayed in console output. Read more about sensitive " +
			"data handling in the [Terraform documentation](https://www.terraform.io/docs/language/state/sensitive-data.html).\n" +
			"\n" +
			"This resource *does* use a cryptographic random number generator.",
		CreateContext: create,
		ReadContext:   readNil,
		DeleteContext: RemoveResourceFromState,
		Schema:        passwordSchema,
		Importer: &schema.ResourceImporter{
			StateContext: importStringFunc(true),
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    resourcePasswordV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourcePasswordStateUpgradeV0,
			},
		},
	}
}

func resourcePasswordV0() *schema.Resource {
	return &schema.Resource{
		Schema: stringSchemaV1(true),
	}
}

func resourcePasswordStateUpgradeV0(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	result, ok := rawState["result"].(string)
	if !ok {
		return nil, fmt.Errorf("resource password state upgrade failed, result could not be asserted as string: %T", rawState["result"])
	}

	hash, err := generateHash(result)
	if err != nil {
		return nil, fmt.Errorf("resource password state upgrade failed, generate hash error: %v", err)
	}

	rawState["bcrypt_hash"] = hash

	return rawState, nil
}

func generateHash(toHash string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(toHash), bcrypt.DefaultCost)

	return string(hash), err
}
