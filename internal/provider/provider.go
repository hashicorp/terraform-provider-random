package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type provider struct {
	configured bool
}

func NewFramework() tfsdk.Provider {
	return &provider{}
}

func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{}, nil
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	p.configured = true
}

func (p *provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"random_id":       resourceIDType{},
		"random_integer":  resourceIntegerType{},
		"random_password": resourcePasswordType{},
		"random_pet":      resourcePetType{},
		"random_shuffle":  resourceShuffleType{},
		"random_string":   resourceStringType{},
		"random_uuid":     resourceUUIDType{},
	}, nil
}

func (p *provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}
