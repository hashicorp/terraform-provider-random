// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/terraform-providers/terraform-provider-random/internal/provider"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/hashicorp/random",
		Debug:   debug,
		// TODO: This should be reverted before merging, a bump to protocol v6 is not necessary to use the new
		// planned_private changes, it just happens that the PR I'm using to test it at the moment only has v6
		// implemented :)
		//
		// https://github.com/hashicorp/terraform/pull/37986
		ProtocolVersion: 6,
	})
	if err != nil {
		log.Fatal(err)
	}
}
