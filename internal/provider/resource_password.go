package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePassword() *schema.Resource {
	return &schema.Resource{
		Create: createStringFunc(true),
		Read:   readNil,
		Delete: schema.RemoveFromState,
		Schema: stringSchemaV1(true),
		Importer: &schema.ResourceImporter{
			StateContext: importStringFunc(true),
		},
	}
}
