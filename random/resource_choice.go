package random

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"math/rand"
	"time"
)

func resourceChoice() *schema.Resource {
	return &schema.Resource{
		Create: CreateChoice,
		Read:   schema.Noop,
		Delete: schema.RemoveFromState,

		Schema: map[string]*schema.Schema{
			"input": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"result": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func CreateChoice(d *schema.ResourceData, _ interface{}) error {
	input := d.Get("input").([]interface{})

	if len(input) < 1 {
		return errors.New("input can't be empty")
	}

	rand.Seed(time.Now().Unix())
	result := input[rand.Intn(len(input))].(string)

	d.SetId("-")
	d.Set("result", result)

	return nil
}
