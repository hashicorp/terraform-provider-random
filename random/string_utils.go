package random

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const numChars = "0123456789"
const lowerChars = "abcdefghijklmnopqrstuvwxyz"
const upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const specialChars = "!@#$%&*()-_=+[]{}<>:?"

func updateStringResourceAttributes(d *schema.ResourceData, charSet string, charClass string, charClassMin string) *schema.ResourceData {
	count := 0
	val := d.Id()
	for i := range val {
		if strings.Contains(charSet, string(val[i])) {
			count = count + 1
		}
	}
	//Check for character set
	if count > 0 {
		d.Set(charClass, true)
	} else {
		d.Set(charClass, false)
	}

	//Check for minimum number of required chars from a character set
	_, newMinClass := d.GetChange(charClassMin)

	if count >= newMinClass.(int) {
		d.Set(charClassMin, newMinClass.(int))
	} else {
		d.Set(charClassMin, count)
	}
	return d

}

func UpperCharsInValue(d *schema.ResourceData) *schema.ResourceData {

	return updateStringResourceAttributes(d, upperChars, "upper", "min_upper")
}

func LowerCharsInValue(d *schema.ResourceData) *schema.ResourceData {
	return updateStringResourceAttributes(d, lowerChars, "lower", "min_lower")
}

func SpecialCharsInValue(d *schema.ResourceData) *schema.ResourceData {
	return updateStringResourceAttributes(d, specialChars, "special", "min_special")
}

func NumbersInValue(d *schema.ResourceData) *schema.ResourceData {
	return updateStringResourceAttributes(d, numChars, "number", "min_numeric")
}
