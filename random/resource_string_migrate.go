package random

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/terraform"
)

func resourceStringMigrateState(v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found Random String State v0; migrating to v1")
		return migrateStringStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateStringStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)

	if v, ok := is.Attributes["min_lower"]; !ok && v == "" {
		is.Attributes["min_lower"] = "0"
	}

	if v, ok := is.Attributes["min_numeric"]; !ok && v == "" {
		is.Attributes["min_numeric"] = "0"
	}

	if v, ok := is.Attributes["min_upper"]; !ok && v == "" {
		is.Attributes["min_upper"] = "0"
	}

	if v, ok := is.Attributes["min_special"]; !ok && v == "" {
		is.Attributes["min_special"] = "0"
	}

	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}
