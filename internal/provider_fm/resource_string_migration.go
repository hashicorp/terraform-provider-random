package provider_fm

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

//func resourceRandomStringMigrateState(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse {
//	switch v {
//	case 0:
//		log.Println("[INFO] Found random string state v0; migrating to v1")
//		return migrateStringStateV0toV1(is)
//	default:
//		return is, fmt.Errorf("Unexpected schema version: %d", v)
//	}
//}

//func resourceRandomStringMigrateState(
//	v int, is *terraform.InstanceState, _ interface{}) (*terraform.InstanceState, error) {
//	switch v {
//	case 0:
//		log.Println("[INFO] Found random string state v0; migrating to v1")
//		return migrateStringStateV0toV1(is)
//	default:
//		return is, fmt.Errorf("Unexpected schema version: %d", v)
//	}
//}

//func redactAttributes(s String) String {
//	s.ID.Value = "<sensitive>"
//	s.Result.Value = "<sensitive>"
//
//	return s
//}

func migrateStringStateV0toV1(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
	s := String{}
	req.State.Get(ctx, &s)
	resp.State.Set(ctx, s)
	if resp.Diagnostics.HasError() {
		return
	}
}
