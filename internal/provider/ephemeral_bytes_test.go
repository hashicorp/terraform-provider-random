// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccEphemeralResourceBytes_basic(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: addEchoBytesConfig(`ephemeral "random_bytes" "test" {
					length = 32
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.bytes_test", tfjsonpath.New("data").AtMapKey("length"), knownvalue.Int64Exact(32)),
					statecheck.ExpectKnownValue("echo.bytes_test", tfjsonpath.New("data").AtMapKey("base64"), knownvalue.StringRegexp(regexp.MustCompile(`^[A-Za-z/+\d]{43}=$`))),
					statecheck.ExpectKnownValue("echo.bytes_test", tfjsonpath.New("data").AtMapKey("hex"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-f\d]{64}$`))),
				},
			},
		},
	})
}

func TestAccEphemeralResourceBytes_Length_ValidationError_AtLeast(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: addEchoBytesConfig(`ephemeral "random_bytes" "test" {
					length = 0
				}`),
				ExpectError: regexp.MustCompile(`Attribute length value must be at least 1, got: 0`),
			},
		},
	})
}

// Adds the test echo provider to enable using state checks with ephemeral resources.
func addEchoBytesConfig(cfg string) string {
	return fmt.Sprintf(`
	%s
	provider "echo" {
		data = ephemeral.random_bytes.test
	}
	resource "echo" "bytes_test" {}
	`, cfg)
}
