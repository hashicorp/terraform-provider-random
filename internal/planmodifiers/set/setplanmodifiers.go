// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package setplanmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
)

func RequiresReplaceIfResultMatchesExclusions() setplanmodifier.RequiresReplaceIfFunc {
	return func(ctx context.Context, req planmodifier.SetRequest, resp *setplanmodifier.RequiresReplaceIfFuncResponse) {
		// TODO: Implement this plan modifier.
		// if existing result matches the exclusion set, enforce recreation of the resource
		// else recreation is not required
		resp.RequiresReplace = true
	}
}
