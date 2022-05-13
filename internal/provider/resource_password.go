package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/crypto/bcrypt"
)

type resourcePasswordType struct{}

func (r resourcePasswordType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	description := "Identical to [random_string](string.html) with the exception that the result is " +
		"treated as sensitive and, thus, _not_ displayed in console output. Read more about sensitive " +
		"data handling in the [Terraform documentation](https://www.terraform.io/docs/language/state/sensitive-data.html).\n" +
		"\n" +
		"This resource *does* use a cryptographic random number generator."

	schema := getStringSchemaV1(true, description)
	schema.Version = 1
	schema.Attributes["bcrypt_hash"] = tfsdk.Attribute{
		Description: "A bcrypt hash of the generated random string.",
		Type:        types.StringType,
		Computed:    true,
		Sensitive:   true,
	}

	return schema, nil
}

func (r resourcePasswordType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourcePassword{
		p: *(p.(*provider)),
	}, nil
}

type resourcePassword struct {
	p provider
}

func (r resourcePassword) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	createPassword(ctx, req, resp)
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r resourcePassword) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}

// Update is intentionally left blank as all required and optional attributes force replacement of the resource
// through the RequiresReplace AttributePlanModifier.
func (r resourcePassword) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r resourcePassword) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

func (r resourcePassword) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	importPassword(ctx, req, resp)
}

func (r resourcePassword) ValidateConfig(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
	validateLength(ctx, req, resp)
}

func (r resourcePassword) UpgradeState(context.Context) map[int64]tfsdk.ResourceStateUpgrader {
	return map[int64]tfsdk.ResourceStateUpgrader{
		0: {
			StateUpgrader: migratePasswordStateV0toV1,
		},
	}
}

func migratePasswordStateV0toV1(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp *tfsdk.UpgradeResourceStateResponse) {
	p := PasswordModel{}
	req.State.Get(ctx, &p)

	hash, err := generateHash(p.Result.Value)
	if err != nil {
		resp.Diagnostics.Append(hashGenerationError(err.Error())...)
		return
	}

	p.BcryptHash = types.String{Value: hash}

	resp.State.Set(ctx, p)
	if resp.Diagnostics.HasError() {
		return
	}
}

func generateHash(toHash string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(toHash), bcrypt.DefaultCost)

	return string(hash), err
}
