package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func New() tfsdk.Provider {
	return &provider{}
}

var _ tfsdk.Provider = (*provider)(nil)

type provider struct{}

func (p *provider) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{}, nil
}

func (p *provider) Configure(context.Context, tfsdk.ConfigureProviderRequest, *tfsdk.ConfigureProviderResponse) {
}

func (p *provider) GetResources(context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"random_id":       &idResourceType{},
		"random_integer":  &integerResourceType{},
		"random_password": &passwordResourceType{},
		"random_pet":      &petResourceType{},
		"random_shuffle":  &shuffleResourceType{},
		"random_string":   &stringResourceType{},
		"random_uuid":     &uuidResourceType{},
	}, nil
}

func (p *provider) GetDataSources(context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}
