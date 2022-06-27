package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

//nolint:unparam
func protoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"random": func() (tfprotov6.ProviderServer, error) {
			return providerserver.NewProtocol6(NewProvider())(), nil
		},
	}
}

func providerVersion332() map[string]resource.ExternalProvider {
	return map[string]resource.ExternalProvider{
		"tls": {
			VersionConstraint: "3.3.2",
			Source:            "hashicorp/random",
		},
	}
}
