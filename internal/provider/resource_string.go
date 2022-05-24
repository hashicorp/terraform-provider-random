package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceString() *schema.Resource {
	return &schema.Resource{
		Description: "The resource `random_string` generates a random permutation of alphanumeric " +
			"characters and optionally special characters.\n" +
			"\n" +
			"This resource *does* use a cryptographic random number generator.\n" +
			"\n" +
			"Historically this resource's intended usage has been ambiguous as the original example used " +
			"it in a password. For backwards compatibility it will continue to exist. For unique ids please " +
			"use [random_id](id.html), for sensitive random values please use [random_password](password.html).",
		CreateContext: createStringFunc(false),
		ReadContext:   readNil,
		DeleteContext: RemoveResourceFromState,
		// MigrateState is deprecated but the implementation is being left in place as per the
		// [SDK documentation](https://github.com/hashicorp/terraform-plugin-sdk/blob/main/helper/schema/resource.go#L91).
		MigrateState:  resourceRandomStringMigrateState,
		SchemaVersion: 1,
		Schema:        stringSchemaV1(),
		Importer: &schema.ResourceImporter{
			StateContext: importStringFunc,
		},
	}
}

func importStringFunc(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	val := d.Id()

	if err := d.Set("result", val); err != nil {
		return nil, fmt.Errorf("error setting result: %w", err)
	}

	return []*schema.ResourceData{d}, nil
}
