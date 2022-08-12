package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func New() provider.Provider {
	return &randomProvider{}
}

var _ provider.Provider = (*randomProvider)(nil)

type randomProvider struct{}

func (p *randomProvider) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{}, nil
}

func (p *randomProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}

func (p *randomProvider) GetResources(context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return map[string]provider.ResourceType{
		"random_id":       &idResourceType{},
		"random_integer":  &integerResourceType{},
		"random_password": &passwordResourceType{},
		"random_pet":      &petResourceType{},
		"random_shuffle":  &shuffleResourceType{},
		"random_string":   &stringResourceType{},
		"random_uuid":     &uuidResourceType{},
	}, nil
}

func (p *randomProvider) GetDataSources(context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{}, nil
}
