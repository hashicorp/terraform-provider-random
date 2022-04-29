package provider_fm

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"testing"
)

func testAccPreCheck(t *testing.T) {
}

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"random": func() (tfprotov6.ProviderServer, error) {
		return providerserver.NewProtocol6(NewFramework())(), nil
	},
}
