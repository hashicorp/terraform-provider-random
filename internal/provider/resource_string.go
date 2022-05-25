package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
		SchemaVersion: 2,
		Schema:        stringSchemaV2(),
		Importer: &schema.ResourceImporter{
			StateContext: importStringFunc,
		},
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 1,
				Type:    resourceStringV1().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceStringStateUpgradeV1,
			},
		},
		CustomizeDiff: customdiff.All(
			customdiff.IfValue("number",
				func(ctx context.Context, value, meta interface{}) bool {
					return !value.(bool)
				},
				func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
					vm := d.GetRawConfig().AsValueMap()
					if vm["number"].IsNull() && vm["numeric"].IsNull() {
						err := d.SetNew("number", true)
						if err != nil {
							return err
						}
						err = d.SetNew("numeric", true)
						if err != nil {
							return err
						}
					}
					return nil
				},
			),
			customdiff.IfValue("numeric",
				func(ctx context.Context, value, meta interface{}) bool {
					return !value.(bool)
				},
				func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
					vm := d.GetRawConfig().AsValueMap()
					if vm["number"].IsNull() && vm["numeric"].IsNull() {
						err := d.SetNew("number", true)
						if err != nil {
							return err
						}
						err = d.SetNew("numeric", true)
						if err != nil {
							return err
						}
					}
					return nil
				},
			),
			customdiff.IfValueChange("number",
				func(ctx context.Context, oldValue, newValue, meta interface{}) bool {
					return oldValue != newValue
				},
				func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
					return d.SetNew("numeric", d.Get("number"))
				},
			),
			customdiff.IfValueChange("numeric",
				func(ctx context.Context, oldValue, newValue, meta interface{}) bool {
					return oldValue != newValue
				},
				func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
					return d.SetNew("number", d.Get("numeric"))
				},
			),
		),
	}
}

func importStringFunc(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	val := d.Id()

	if err := d.Set("result", val); err != nil {
		return nil, fmt.Errorf("error setting result: %w", err)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceStringV1() *schema.Resource {
	return &schema.Resource{
		Schema: stringSchemaV1(),
	}
}

func resourceStringStateUpgradeV1(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	return resourceStateUpgradeAddNumeric(ctx, rawState, meta, "string")(ctx, rawState, meta)
}
