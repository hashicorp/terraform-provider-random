package random

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceRange() *schema.Resource {
	return &schema.Resource{
		Create: CreateRange,
		Read:   RepopulateRange,
		Delete: schema.RemoveFromState,

		Schema: map[string]*schema.Schema{
			"keepers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"min": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"max": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"seed": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func CreateRange(d *schema.ResourceData, meta interface{}) error {
	min := d.Get("min").(int)
	max := d.Get("max").(int)
	seed := d.Get("seed").(string)

	if max <= min {
		return fmt.Errorf("Minimum value needs to be smaller than maximum value")
	}
	rand := NewRand(seed)
	number := rand.Intn((max+1)-min) + min
	d.SetId(strconv.Itoa(number))

	return nil
}

func RepopulateRange(d *schema.ResourceData, _ interface{}) error {
	return nil
}
