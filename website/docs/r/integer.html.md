---
layout: "random"
page_title: "Random: random_integer"
sidebar_current: "docs-random-resource-integer"
description: |-
  Generates a random integer values.
---

# random\_integer

The resource `random_integer` generates random values from a given range, described by the `min` and `max` attributes of a given resource.

This resource can be used in conjunction with resources that have
the `create_before_destroy` lifecycle flag set, to avoid conflicts with
unique names during the brief period where both the old and new resources
exist concurrently.

## Example Usage

The following example shows how to generate a random priority between 1 and 99999 for
a `aws_alb_listener_rule` resource:

```hcl
resource "random_integer" "priority" {
  min     = 1
  max     = 99999
  keepers = {
    # Generate a new integer each time we switch to a new listener ARN
    listener_arn = "${var.listener_arn}"
  }
}

resource "aws_alb_listener_rule" "main" {
  listener_arn = "${var.listener_arn}"
  priority     = "${random_integer.priority.result}"

  action {
    type             = "forward"
    target_group_arn = "${var.target_group_arn}"
  }
  # ... (other aws_alb_listener_rule arguments) ...
}
```

The result of the above will set a random priority.

## Argument Reference

The following arguments are supported:

* `min` - (int) The minimum inclusive value of the range.

* `max` - (int) The maximum inclusive value of the range.

* `keepers` - (Optional) Arbitrary map of values that, when changed, will
  trigger a new id to be generated. See
  [the main provider documentation](../index.html) for more information.

* `seed` - (Optional) A custom seed to always produce the same value.

## Attribute Reference

The following attributes are supported:

* `id` - (string) An internal id.
* `result` - (int) The random Integer result.

## Import

Random integers can be imported using the `result`, `min`, and `max`, with an optional `seed`.
This can be used to replace a config value with a value interpolated from the random provider without experiencing diffs.

Example (values are separated by a `,`):
```
$ terraform import random_integer.priority 15390,1,99999
```
