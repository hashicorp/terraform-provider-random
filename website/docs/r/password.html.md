---
layout: "random"
page_title: "Random: random_password"
sidebar_current: "docs-random-resource-password"
description: |-
  Identical to random_string with the exception that the result is treated as sensitive.
---

# random\_password

Identical to [random_string](string.html) with the exception that the
result is treated as sensitive and, thus, not displayed in console output.

This resource *does* use a cryptographic random number generator.

## Example Usage

```hcl
resource "random_password" "password" {
  length = 16
  special = true
  override_special = "/@\" "
}

resource "aws_db_instance" "example" {
  password = "${random_password.password.result}"
}
```
