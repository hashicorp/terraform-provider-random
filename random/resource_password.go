package random

import (
	"github.com/hashicorp/terraform/helper/schema"
)

/*********************************************************
  See resource_string.go for the implementation.
  resource_password and resource_string are intended to
  be identical other than the result of resource_password
  is treated as sensitive information.
*********************************************************/

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
