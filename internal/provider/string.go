// This file provides shared functionality between `resource_string` and `resource_password`.
// There is no intent to permanently couple their implementations
// Over time they could diverge, or one becomes deprecated
package provider

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func stringSchemaV1(sensitive bool) map[string]*schema.Schema {
	idDesc := "The generated random string."
	if sensitive {
		idDesc = "A static value used internally by Terraform, this should not be referenced in configurations."
	}

	return map[string]*schema.Schema{
		"keepers": {
			Description: "Arbitrary map of values that, when changed, will trigger recreation of " +
				"resource. See [the main provider documentation](../index.html) for more information.",
			Type:     schema.TypeMap,
			Optional: true,
			ForceNew: true,
		},

		"length": {
			Description: "The length of the string desired.",
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
		},

		"special": {
			Description: "Include special characters in the result. These are `!@#$%&*()-_=+[]{}<>:?`",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			ForceNew:    true,
		},

		"upper": {
			Description: "Include uppercase alphabet characters in the result.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			ForceNew:    true,
		},

		"lower": {
			Description: "Include lowercase alphabet characters in the result.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			ForceNew:    true,
		},

		"number": {
			Description: "Include numeric characters in the result.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			ForceNew:    true,
		},

		"min_numeric": {
			Description: "Minimum number of numeric characters in the result.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			ForceNew:    true,
		},

		"min_upper": {
			Description: "Minimum number of uppercase alphabet characters in the result.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			ForceNew:    true,
		},

		"min_lower": {
			Description: "Minimum number of lowercase alphabet characters in the result.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			ForceNew:    true,
		},

		"min_special": {
			Description: "Minimum number of special characters in the result.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			ForceNew:    true,
		},

		"override_special": {
			Description: "Supply your own list of special characters to use for string generation.  This " +
				"overrides the default character list in the special argument.  The `special` argument must " +
				"still be set to true for any overwritten characters to be used in generation.",
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},

		"result": {
			Description: "The generated random string.",
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   sensitive,
		},

		"id": {
			Description: idDesc,
			Computed:    true,
			Type:        schema.TypeString,
		},
	}
}

func createStringFunc(sensitive bool) func(d *schema.ResourceData, meta interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		const numChars = "0123456789"
		const lowerChars = "abcdefghijklmnopqrstuvwxyz"
		const upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		var specialChars = "!@#$%&*()-_=+[]{}<>:?"

		length := d.Get("length").(int)
		upper := d.Get("upper").(bool)
		minUpper := d.Get("min_upper").(int)
		lower := d.Get("lower").(bool)
		minLower := d.Get("min_lower").(int)
		number := d.Get("number").(bool)
		minNumeric := d.Get("min_numeric").(int)
		special := d.Get("special").(bool)
		minSpecial := d.Get("min_special").(int)
		overrideSpecial := d.Get("override_special").(string)

		if overrideSpecial != "" {
			specialChars = overrideSpecial
		}

		var chars = string("")
		if upper {
			chars += upperChars
		}
		if lower {
			chars += lowerChars
		}
		if number {
			chars += numChars
		}
		if special {
			chars += specialChars
		}

		minMapping := map[string]int{
			numChars:     minNumeric,
			lowerChars:   minLower,
			upperChars:   minUpper,
			specialChars: minSpecial,
		}
		var result = make([]byte, 0, length)
		for k, v := range minMapping {
			s, err := generateRandomBytes(&k, v)
			if err != nil {
				return errwrap.Wrapf("error generating random bytes: {{err}}", err)
			}
			result = append(result, s...)
		}
		s, err := generateRandomBytes(&chars, length-len(result))
		if err != nil {
			return errwrap.Wrapf("error generating random bytes: {{err}}", err)
		}
		result = append(result, s...)
		order := make([]byte, len(result))
		if _, err := rand.Read(order); err != nil {
			return errwrap.Wrapf("error generating random bytes: {{err}}", err)
		}
		sort.Slice(result, func(i, j int) bool {
			return order[i] < order[j]
		})

		d.Set("result", string(result))
		if sensitive {
			d.SetId("none")
		} else {
			d.SetId(string(result))
		}
		return nil
	}
}

func generateRandomBytes(charSet *string, length int) ([]byte, error) {
	bytes := make([]byte, length)
	setLen := big.NewInt(int64(len(*charSet)))
	for i := range bytes {
		idx, err := rand.Int(rand.Reader, setLen)
		if err != nil {
			return nil, err
		}
		bytes[i] = (*charSet)[idx.Int64()]
	}
	return bytes, nil
}

func readNil(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func importStringFunc(sensitive bool) schema.StateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
		val := d.Id()
		if sensitive {
			d.SetId("none")
		}
		// common string passed to guess parameters based on the schema
		// "password keepers={\"bla\": \"dibla\",\"key\":\"value\"},length=25,special=true,upper=true,lower=true,number=true,min_numeric=0,min_upper=0,min_lower=0,min_special=0,override_special=_%@"
		// use a regex to fetch all settings
		// Note: override_special should be the last specified key if specified
		settings := map[string]interface{}{
			// At some point we should use a default value here, why not 16 then.
			"keepers":          nil,
			"length":           16,
			"special":          true,
			"upper":            true,
			"lower":            true,
			"number":           true,
			"min_numeric":      0,
			"min_upper":        0,
			"min_lower":        0,
			"min_special":      0,
			"override_special": "",
		}

		exploded := strings.Fields(val)

		switch len(exploded) {
		case 1:
			// Keep default behavior but add default values
			d.Set("result", val)
			log.Println(fmt.Sprintf("[DEBUG] using default settings: %v", settings))
			for key, value := range settings {
				d.Set(key, value)
			}
			if !sensitive {
				d.SetId(val)
			}
		case 2:
			d.Set("result", exploded[0])
			if !sensitive {
				d.SetId(exploded[0])
			}
			settingsRegexp := regexp.MustCompile(`(\w+)=({.*}|nil|\d+|true|false|nil|[[:ascii:]]+)`)
			extractedSettings := settingsRegexp.FindAllStringSubmatch(exploded[1], -1)
			// Setting might need to be casted to int8, to a map[string]interface{} (from jsonString) or a bool or
			// a literal nil
			for _, setting := range extractedSettings {
				asInt, err := strconv.Atoi(setting[2])
				if err != nil {
					// Test for a bool
					parsedBool, err := strconv.ParseBool(setting[2])
					if err != nil {
						// Test for jsonString
						var jsonMap map[string]interface{}
						err = json.Unmarshal([]byte(setting[2]), &jsonMap)
						if err != nil {
							// it can only be a string then
							if setting[2] == "nil" {
								log.Println(fmt.Sprintf("[DEBUG] Adding setting %v with value %v valuetype nil", setting[1], nil))
								settings[setting[1]] = nil
							} else {
								log.Println(fmt.Sprintf("[DEBUG] Adding setting %v with value %v valuetype string", setting[1], setting[2]))
								settings[setting[1]] = setting[2]
							}
						} else {
							log.Println(fmt.Sprintf("[DEBUG] Adding setting %v with value %v valuetype map[string]interface{}", setting[1], jsonMap))
							settings[setting[1]] = jsonMap
						}
					} else {
						log.Println(fmt.Sprintf("[DEBUG] Adding setting %v with value %v valuetype bool", setting[1], parsedBool))
						settings[setting[1]] = parsedBool
					}
				} else {
					log.Println(fmt.Sprintf("[DEBUG] Adding setting %v with value %v valuetype int8", setting[1], asInt))
					settings[setting[1]] = asInt
				}
			}
			log.Println(fmt.Sprintf("[DEBUG] new settings: %v", settings))
			for key, value := range settings {
				d.Set(key, value)
			}

		default:
			return nil, fmt.Errorf("unable to parse string `%s` and extract string or string and settings\n"+
				"ressource address should be in the following form:\n\"password key=value,key2=value2...\"", val)

		}
		return []*schema.ResourceData{d}, nil
	}
}
