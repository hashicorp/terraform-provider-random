---
layout: "random"
page_title: "Random: random_string"
sidebar_current: "docs-random-resource-string"
description: |-
  Produces a random string of a length using alphanumeric characters and optionally special characters.
---

# random\_string

The resource `random_string` generates a random permutation of alphanumeric
characters and optionally special characters.

This resource *does* use a cryptographic random number generator.

## Example Usage

```hcl
resource "random_string" "password" {
  length = 16
  special = true
  override_special = "/@\" "
}

resource "aws_db_instance" "example" {
  password = "${random_string.password.result}"
}
```

## Argument Reference

The following arguments are supported:

* `length` - (Required) The length of the string desired

* `upper` - (Optional) (default true) Include uppercase alphabet characters
  in random string.

* `min_upper` - (Optional) (default 0) Minimum number of uppercase alphabet
  characters in random string.

* `lower` - (Optional) (default true) Include lowercase alphabet characters
  in random string.

* `min_lower` - (Optional) (default 0) Minimum number of lowercase alphabet
  characters in random string.

* `number` - (Optional) (default true) Include numeric characters in random
  string.

* `min_numeric` - (Optional) (default 0) Minimum number of numeric characters
  in random string.

* `special` - (Optional) (default true) Include special characters in random
  string. These are '!@#$%&*()-_=+[]{}<>:?'

* `min_special` - (Optional) (default 0) Minimum number of special characters
  in random string.

* `override_special` - (Optional) Supply your own list of special characters to
  use for string generation.  This overrides characters list in the special
  argument.  The special argument must still be set to true for any overwritten
  characters to be used in generation.

* `keepers` - (Optional) Arbitrary map of values that, when changed, will
  trigger a new id to be generated. See
  [the main provider documentation](../index.html) for more information.


## Attributes Reference

The following attributes are exported:

* `result` - Random string generated.

