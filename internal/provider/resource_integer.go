package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceInteger() *schema.Resource {
	return &schema.Resource{
		Description: "The resource `random_integer` generates random values from a given range, described " +
			"by the `min` and `max` attributes of a given resource.\n" +
			"\n" +
			"This resource can be used in conjunction with resources that have the `create_before_destroy` " +
			"lifecycle flag set, to avoid conflicts with unique names during the brief period where both the " +
			"old and new resources exist concurrently.",
		CreateContext: CreateInteger,
		ReadContext:   schema.NoopContext,
		DeleteContext: RemoveResourceFromState,
		Importer: &schema.ResourceImporter{
			StateContext: ImportInteger,
		},

		Schema: map[string]*schema.Schema{
			"keepers": {
				Description: "Arbitrary map of values that, when changed, will trigger recreation of " +
					"resource. See [the main provider documentation](../index.html) for more information.",
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},

			"min": {
				Description: "The minimum inclusive value of the range.",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},

			"max": {
				Description: "The maximum inclusive value of the range.",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},

			"seed": {
				Description: "A custom seed to always produce the same value.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},

			"result": {
				Description: "The random integer result.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"id": {
				Description: "The string representation of the integer result.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
		UseJSONNumber: true,
	}
}

func CreateInteger(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	min := d.Get("min").(int)
	max := d.Get("max").(int)
	seed := d.Get("seed").(string)

	if max <= min {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "minimum value needs to be smaller than maximum value",
		})
	}
	rand := NewRand(seed)
	number := rand.Intn((max+1)-min) + min

	if err := d.Set("result", number); err != nil {
		return diag.Errorf("error setting result: %s", err)
	}

	d.SetId(strconv.Itoa(number))

	return nil
}

func ImportInteger(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ",")
	if len(parts) != 3 && len(parts) != 4 {
		return nil, fmt.Errorf("Invalid import usage: expecting {result},{min},{max} or {result},{min},{max},{seed}")
	}

	result, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("error parsing result: %w", err)
	}

	if err := d.Set("result", result); err != nil {
		return nil, fmt.Errorf("error setting result: %w", err)
	}

	min, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing min: %w", err)
	}

	if err := d.Set("min", min); err != nil {
		return nil, fmt.Errorf("error setting min: %w", err)
	}

	max, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("error parsing max: %w", err)
	}

	if err := d.Set("max", max); err != nil {
		return nil, fmt.Errorf("error setting max: %w", err)
	}

	if len(parts) == 4 {
		if err := d.Set("seed", parts[3]); err != nil {
			return nil, fmt.Errorf("error setting seed: %w", err)
		}
	}

	d.SetId(parts[0])

	return []*schema.ResourceData{d}, nil
}
