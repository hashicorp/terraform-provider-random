---
layout: "random"
page_title: "Random: random_password"
sidebar_current: "docs-random-resource-password"
description: |-
  Produces a random string of a length using alphanumeric characters and optionally special characters. The result will not be displayed to console.
---

# random\_password

~> **Note:** Requires random provider version >= 2.2.0

Identical to [random_string](string.html) with the exception that the
result is treated as sensitive and, thus, _not_ displayed in console output.
It also adds the additional field `bcrypt_hash`.

~> **Note:** All attributes including the generated password will be stored in
the raw state as plain-text. [Read more about sensitive data in
state](/docs/state/sensitive-data.html).

This resource *does* use a cryptographic random number generator.

## Example Usage

```hcl
resource "random_password" "password" {
  length = 16
  special = true
  override_special = "_%@"
}

resource "aws_db_instance" "example" {
  instance_class = "db.t3.micro"
  allocated_storage = 64
  engine = "mysql"
  username = "someone"
  password = random_password.password.result
}
```

## Import

Random Password can be imported by specifying the value of the string:

```
terraform import random_password.password securepassword
```

## Argument Reference

random_password has the same arguments as listed in [random_string](string.html).

## Attribute Reference

The following attribute is added to [random_string](string.html):

* `bcrypt_hash` - (string) The random string hashed with the bcrypt adaptive
  hashing algorithm. Hashing with bcrypt always produces a different result
  (even for the same input), which can be annoying to use. This attribute will
  only change when another random string is generated.
