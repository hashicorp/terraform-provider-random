---
page_title: "random_bytes Resource - terraform-provider-random"
subcategory: ""
description: |-
  The resource random_bytes generates an array of bytes that is intended to be used as key or secret.
---

# random_bytes (Resource)

The resource `random_bytes` generates an array of bytes that is intended to be used as key or secret.

This resource *does* use a cryptographic random number generator.

## Example Usage

```terraform
resource "random_bytes" "secret" {
  length           = 64
}

resource "azurerm_key_vault_secret" "jwt_secret" {
  key_vault_id = "some-azure-key-vault-id"
  name         = "JwtSecret"
  value        = random_bytes.jwt_secret.result_base64
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `length` (Number) The number of bytes to generate.

### Optional

- `keepers` (Map of String) Arbitrary map of values that, when changed, will trigger recreation of resource. See [the main provider documentation](../index.html) for more information.

### Read-Only

- `result_base64` (String, Sensitive) The generated bytes presented in base64 string format.
- `result_hex` (String, Sensitive) The generated bytes presented in hex string format.
- `id` (String) A static value used internally by Terraform, this should not be referenced in configurations.

## Import

Import is supported using the following syntax with the bytes encoded in base64:

```shell
terraform import random_bytes.secret 8/fu3q+2DcgSJ19i0jZ5Cw==
```