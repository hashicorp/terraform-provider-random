package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/terraform-providers/terraform-provider-random/random"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: random.Provider})
}
