# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Random UUID's can be imported. This can be used to replace a config
# value with a value interpolated from the random provider without
# experiencing diffs.

terraform import random_uuid.main aabbccdd-eeff-0011-2233-445566778899