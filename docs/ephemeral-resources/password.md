---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "random_password Ephemeral Resource - terraform-provider-random"
subcategory: ""
description: |-
  Generates an ephemeral password string using a cryptographic random number generator.
  A random ephemeral password used in combination with a write-only resource attribute will avoid Terraform storing the password string in the plan or state file.
---

# random_password (Ephemeral Resource)

Generates an ephemeral password string using a cryptographic random number generator.

A random ephemeral password used in combination with a write-only resource attribute will avoid Terraform storing the password string in the plan or state file.

## Example Usage

```terraform
ephemeral "random_password" "password" {
  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `length` (Number) The length of the string desired. The minimum value for length is 1 and, length must also be >= (`min_upper` + `min_lower` + `min_numeric` + `min_special`).

### Optional

- `lower` (Boolean) Include lowercase alphabet characters in the result. Default value is `true`.
- `min_lower` (Number) Minimum number of lowercase alphabet characters in the result. Default value is `0`.
- `min_numeric` (Number) Minimum number of numeric characters in the result. Default value is `0`.
- `min_special` (Number) Minimum number of special characters in the result. Default value is `0`.
- `min_upper` (Number) Minimum number of uppercase alphabet characters in the result. Default value is `0`.
- `numeric` (Boolean) Include numeric characters in the result. Default value is `true`. If `numeric`, `upper`, `lower`, and `special` are all configured, at least one of them must be set to `true`.
- `override_special` (String) Supply your own list of special characters to use for string generation.  This overrides the default character list in the special argument.  The `special` argument must still be set to true for any overwritten characters to be used in generation.
- `special` (Boolean) Include special characters in the result. These are `!@#$%&*()-_=+[]{}<>:?`. Default value is `true`.
- `upper` (Boolean) Include uppercase alphabet characters in the result. Default value is `true`.

### Read-Only

- `bcrypt_hash` (String, Sensitive) A bcrypt hash of the generated random string. **NOTE**: If the generated random string is greater than 72 bytes in length, `bcrypt_hash` will contain a hash of the first 72 bytes.
- `result` (String, Sensitive) The generated random string.