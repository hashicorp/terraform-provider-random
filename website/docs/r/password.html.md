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

~> **Note:** All attributes including the generated password will be stored in
the raw state as plain-text. [Read more about sensitive data in
state](/docs/state/sensitive-data.html).

This resource *does* use a cryptographic random number generator.

## Example Usage

```hcl
resource "random_password" "password" {
  length           = 30
  override_special = "_%@"
  min_lower        = 5
  min_numeric      = 5
  min_special      = 6
  min_upper        = 9
}


resource "aws_db_instance" "example" {
  instance_class = "db.t3.micro"
  allocated_storage = 64
  engine = "mysql"
  username = "someone"
  password = random_password.password.result
}
```


## Argument Reference

The following arguments are supported:
* `length` - (Required) The desired length of the password.

* `special` - (Optional) (boolean) Include special characters in random
  string. Default is `true`. Default special characters included are `!@#$%&*()-_=+[]{}<>:?`
  
* `min_special` - Sets the minimum amount of special characters to be included in password. Chooses from default or override list as applicable.

* `override_special` - (Optional) Supply your own list of special characters to
  use for string generation.  This overrides the default character list in the *special*
  argument. `WARNING:` The *special* argument must still be set to `true` for any overwritten
  characters to be used in generation, else using this argument may crash terraform execution.

* `number` - (Optional) (boolean) Specifies if numbers should be included in the password string. 
             Default is `true`.

* `min_numeric` - (Optional) (number) Sets the minimum amount of digits to be included in password. 
                  Overrides *number = false*. Default value is 0

* `upper` - (Optional) (boolean) Specifies if upper case alphabets should be included in the password string. 
             Default is `true`.

* `min_upper` - (Optional) (number) Sets the minimum amount of upper case alphabets to be included in password. 
                Overrides *upper = false*. Default value is 0

* `lower` - (Optional) (boolean) Specifies if lower case alphabets should be included in the password string. 
            Default is `true`.

* `min_lower` - (Optional) (number) Sets the minimum amount of lower case alphabets to be included in password. 
                Overrides *lower = false*. Default value is 0

* `keepers` - (Optional) Arbitrary map of values that, when changed, will
  trigger a new id to be generated. See
  [the main provider documentation](../index.html) for more information.


## Attribute Reference

The following attributes are exported:

* `result` - Random password generated.

* `id` - The internal id of the resource. `None` is the default value.

## Import

Random Password can be imported by specifying the value of the string:

```
terraform import random_password.password securepassword
```
