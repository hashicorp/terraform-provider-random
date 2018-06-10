---
layout: "random"
page_title: "Random: random_choice"
sidebar_current: "docs-random-resource-choice"
description: |-
  Produces a random element of a given list.
---

# random\_choice

The resource `random_choice` selects a random element from a list
of strings given as an argument.

## Example Usage

```hcl
resource "random_choice" "az" {
  input = ["us-west-1a", "us-west-1c", "us-west-1d", "us-west-1e"]
}

resource "aws_elb" "example" {
  # Place the ELB in a randomly selected availability zone
  availability_zones = ["${random_choice.az.result}"]

  # ... and other aws_elb arguments ...
}
```

## Argument Reference

The following arguments are supported:

* `input` - (Required) The non-empty list of strings.

## Attributes Reference

The following attributes are exported:

* `result` - Random element selected from the list of strings given in `input`.

