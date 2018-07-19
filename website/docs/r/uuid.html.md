---
layout: "random"
page_title: "Random: random_uuid"
sidebar_current: "docs-random-resource-id"
description: |-
  Generates a random identifier.
---

# random\_uuid

The resource `random_uuid` generates random uuid string that is intended to be
used as unique identifiers for other resources.

This resource uses the `hashicorp/go-uuid` to generate a UUID-formatted string
for use with services needed a unique string identifier.


## Example Usage

The following example shows how to generate a unique name for an Azure Resource Group.

```hcl
resource "random_uuid" "test" { }

resource "azurerm_resource_group" "test" {
  name     = "${random_uuid.test.result}-rg"
  location = "Central US"
}
```

## Argument Reference

The following arguments are supported:

* `keepers` - (Optional) Arbitrary map of values that, when changed, will
  trigger a new uuid to be generated. See
  [the main provider documentation](../index.html) for more information.

## Attributes Reference

The following attributes are exported:

* `result` - The generated uuid presented in string format.

## Import

Random UUID's can be imported. This can be used to replace a config value with a value
interpolated from the random provider without experiencing diffs.

Example:
```
$ terraform import random_uuid.main aabbccdd-eeff-0011-2233-445566778899
```
