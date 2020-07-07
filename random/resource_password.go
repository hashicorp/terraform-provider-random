package random

import (
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"golang.org/x/crypto/bcrypt"
)

func resourcePassword() *schema.Resource {
	passwordSchema := stringSchemaV1(true)
	passwordSchema["bcrypt_hash"] = &schema.Schema{
		Type:      schema.TypeString,
		Computed:  true,
		Sensitive: true,
	}

	create := func(d *schema.ResourceData, meta interface{}) error {
		if err := createStringFunc(true)(d, meta); err != nil {
			return err
		}
		if d.Get("bcrypt_hash") == "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(d.Get("result").(string)), bcrypt.DefaultCost)
			if err != nil {
				return errwrap.Wrapf("error hashing random bytes: {{err}}", err)
			}
			d.Set("bcrypt_hash", string(hash))
		}
		return nil
	}

	return &schema.Resource{
		Create: create,
		Read:   readNil,
		Delete: schema.RemoveFromState,
		Schema: passwordSchema,
		Importer: &schema.ResourceImporter{
			State: importStringFunc(true),
		},
	}
}
