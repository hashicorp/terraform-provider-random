package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-provider-random/random"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: random.Provider})
}
