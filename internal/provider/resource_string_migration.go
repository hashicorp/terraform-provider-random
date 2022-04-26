package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func resourceRandomStringMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {

	ctx := context.TODO()

	switch v {
	case 0:
		tflog.Info(ctx, "Found random string state v0; migrating to v1")
		return migrateStringStateV0toV1(ctx, is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func redactAttributes(is *terraform.InstanceState) map[string]string {
	redactedAttributes := make(map[string]string)
	for k, v := range is.Attributes {
		redactedAttributes[k] = v
		if k == "id" || k == "result" {
			redactedAttributes[k] = "<sensitive>"
		}
	}
	return redactedAttributes
}

func migrateStringStateV0toV1(ctx context.Context, is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() {
		tflog.Debug(ctx, "Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	tflog.Debug(ctx, fmt.Sprintf("Random String Attributes before Migration: %#v", redactAttributes(is)))

	keys := []string{"min_numeric", "min_upper", "min_lower", "min_special"}
	for _, k := range keys {
		if v := is.Attributes[k]; v == "" {
			is.Attributes[k] = "0"
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Random String Attributes after State Migration: %#v", redactAttributes(is)))

	return is, nil
}
