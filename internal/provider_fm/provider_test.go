package provider_fm

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf6to5server"
	"log"
	"testing"
)

func testAccPreCheck(t *testing.T) {
}

//nolint:unparam
func testAccProtoV5ProviderFactories() map[string]func() (tfprotov5.ProviderServer, error) {
	downgradedFrameworkProvider, err := tf6to5server.DowngradeServer(context.Background(), func() tfprotov6.ProviderServer {
		return providerserver.NewProtocol6(NewFramework())()
	})
	if err != nil {
		log.Fatal(err)
	}

	return map[string]func() (tfprotov5.ProviderServer, error){
		"random": func() (tfprotov5.ProviderServer, error) {
			return downgradedFrameworkProvider, nil
		},
	}
}
