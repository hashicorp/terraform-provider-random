package random

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourcePassword() *schema.Resource {
	resourcePassword := resourceString()
	resourcePassword.Create = CreatePassword
	resourcePassword.Schema["result"].Sensitive = true
	return resourcePassword
}

func CreatePassword(d *schema.ResourceData, meta interface{}) (err error) {
	if err = CreateString(d, meta); err != nil {
		return err
	}
	d.SetId("none")
	return nil
}
