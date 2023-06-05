# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "random_string" "random" {
  length           = 16
  special          = true
  override_special = "/@£$"
}
