package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/terraform-providers/terraform-provider-random/internal/resources/id"
	"github.com/terraform-providers/terraform-provider-random/internal/resources/integer"
	"github.com/terraform-providers/terraform-provider-random/internal/resources/password"
	"github.com/terraform-providers/terraform-provider-random/internal/resources/pet"
	"github.com/terraform-providers/terraform-provider-random/internal/resources/shuffle"
	"github.com/terraform-providers/terraform-provider-random/internal/resources/stringresource"
	"github.com/terraform-providers/terraform-provider-random/internal/resources/uuid"
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
		"random_id":       id.NewResourceType(),
		"random_integer":  integer.NewResourceType(),
		"random_password": password.NewResourceType(),
		"random_pet":      pet.NewResourceType(),
		"random_shuffle":  shuffle.NewResourceType(),
		"random_string":   stringresource.NewResourceType(),
		"random_uuid":     uuid.NewResourceType(),
	}, nil
}

func (p *provider) GetDataSources(context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}
