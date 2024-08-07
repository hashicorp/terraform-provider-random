// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package randomtest

import (
	"context"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"

	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var _ statecheck.StateCheck = expectNoAttribute{}

type expectNoAttribute struct {
	resourceAddress string
	attributePath   tfjsonpath.Path
}

// CheckState implements the state check logic.
func (e expectNoAttribute) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	var resource *tfjson.StateResource

	if req.State == nil {
		resp.Error = fmt.Errorf("state is nil")

		return
	}

	if req.State.Values == nil {
		resp.Error = fmt.Errorf("state does not contain any state values")

		return
	}

	if req.State.Values.RootModule == nil {
		resp.Error = fmt.Errorf("state does not contain a root module")

		return
	}

	for _, r := range req.State.Values.RootModule.Resources {
		if e.resourceAddress == r.Address {
			resource = r

			break
		}
	}

	if resource == nil {
		resp.Error = fmt.Errorf("%s - Resource not found in state", e.resourceAddress)

		return
	}

	_, err := tfjsonpath.Traverse(resource.AttributeValues, e.attributePath)

	// Attribute doesn't exist in the resource, which is the success scenario for this state check.
	if err != nil {
		return
	}

	resp.Error = fmt.Errorf("%s - Attribute %q was found in resource state, but was expected to not exist", e.resourceAddress, e.attributePath.String())
}

// ExpectNoAttribute returns a state check that asserts that the specified attribute at the given resource does not exist.
func ExpectNoAttribute(resourceAddress string, attributePath tfjsonpath.Path) statecheck.StateCheck {
	return expectNoAttribute{
		resourceAddress: resourceAddress,
		attributePath:   attributePath,
	}
}
