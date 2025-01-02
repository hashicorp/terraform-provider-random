// Copyright (c) HashiCorp, Inc.
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
	"github.com/terraform-providers/terraform-provider-random/internal/randomtest"
)

func TestAccEphemeralResourcePassword_basic(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		// Ephemeral resources are only available in 1.10 and later
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: addEchoConfig(`ephemeral "random_password" "test" {
					length = 20
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("result"), randomtest.StringLengthExact(20)),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("special"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("upper"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("lower"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("numeric"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("min_numeric"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("min_upper"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("min_lower"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("min_special"), knownvalue.Int64Exact(0)),
				},
			},
		},
	})
}

func TestAccEphemeralResourcePassword_BcryptHash(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		// Ephemeral resources are only available in 1.10 and later
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: addEchoConfig(`ephemeral "random_password" "test" {
					length = 73
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						"echo.password_test", tfjsonpath.New("data").AtMapKey("bcrypt_hash"),
						"echo.password_test", tfjsonpath.New("data").AtMapKey("result"),
						randomtest.BcryptHashMatch(),
					),
				},
			},
		},
	})
}

func TestAccEphemeralResourcePassword_Override(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		// Ephemeral resources are only available in 1.10 and later
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: addEchoConfig(`ephemeral "random_password" "test" {
					length = 4
					override_special = "!"
					lower = false
					upper = false
					numeric = false
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("result"), randomtest.StringLengthExact(4)),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("result"), knownvalue.StringExact("!!!!")),
				},
			},
		},
	})
}

func TestAccEphemeralResourcePassword_Min(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		// Ephemeral resources are only available in 1.10 and later
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: addEchoConfig(`ephemeral "random_password" "test" {
					length = 12
					override_special = "!#@"
					min_lower = 2
					min_upper = 3
					min_special = 1
					min_numeric = 4
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("result"), randomtest.StringLengthExact(12)),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("result"), knownvalue.StringRegexp(regexp.MustCompile(`([a-z].*){2,}`))),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("result"), knownvalue.StringRegexp(regexp.MustCompile(`([A-Z].*){3,}`))),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("result"), knownvalue.StringRegexp(regexp.MustCompile(`([0-9].*){4,}`))),
					statecheck.ExpectKnownValue("echo.password_test", tfjsonpath.New("data").AtMapKey("result"), knownvalue.StringRegexp(regexp.MustCompile(`([!#@])`))),
				},
			},
		},
	})
}

func TestAccEphemeralResourcePassword_Numeric_ValidationError(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		// Ephemeral resources are only available in 1.10 and later
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: addEchoConfig(`ephemeral "random_password" "test" {
					length = 12
					special = false
					upper = false
					lower = false
					numeric = false
				}`),
				ExpectError: regexp.MustCompile(`At least one attribute out of \[special,upper,lower,numeric\] must be specified`),
			},
		},
	})
}

func TestAccEphemeralResourcePassword_Length_ValidationError_SumOf(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		// Ephemeral resources are only available in 1.10 and later
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: addEchoConfig(`ephemeral "random_password" "test" {
					length		= 11
					min_upper	= 3
					min_lower	= 3
					min_numeric	= 3
					min_special	= 3
				}`),
				ExpectError: regexp.MustCompile(`Attribute length value must be at least sum of`),
			},
		},
	})
}

func TestAccEphemeralResourcePassword_Length_ValidationError_AtLeast(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		// Ephemeral resources are only available in 1.10 and later
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: addEchoConfig(`ephemeral "random_password" "test" {
					length		= 0
				}`),
				ExpectError: regexp.MustCompile(`Attribute length value must be at least 1, got: 0`),
			},
		},
	})
}

// Adds the test echo provider to enable using state checks with ephemeral resources.
func addEchoConfig(cfg string) string {
	return fmt.Sprintf(`
	%s
	provider "echo" {
		data = ephemeral.random_password.test
	}
	resource "echo" "password_test" {}
	`, cfg)
}
