## 2.0.0 (August 15, 2018)

BACKWARDS INCOMPATIBILITIES / NOTES:
* `random_string`: set the ID for random_string resources to "none". Any terraform configuration referring to `random_string.foo.id` will need to be updated to reference `random_string.foo.result` ([#17](https://github.com/terraform-providers/terraform-provider-random/issues/17))

NEW FEATURES:

* `random_uuid` generates random uuid string that is intended to be used as unique identifiers for other resources ([#38](https://github.com/terraform-providers/terraform-provider-random/issues/38))

BUG FIXES: 
* Use UnixNano() instead of Unix() for the current time seed in NewRand() ([#27](https://github.com/terraform-providers/terraform-provider-random/issues/27))
* `random_shuffle`: if `random_shuffle` is given an empty list, it will return an empty list

IMPROVEMENTS:

* Replace ReadPet function in `resource_pet` with schema.Noop ([#34](https://github.com/terraform-providers/terraform-provider-random/issues/34))

## 1.3.1 (May 22, 2018)

BUG FIXES:

* Add migration and new schema version for `resource_string` ([#29](https://github.com/terraform-providers/terraform-provider-random/issues/29))

## 1.3.0 (May 21, 2018)

BUG FIXES:

* `random_integer` now supports update ([#25](https://github.com/terraform-providers/terraform-provider-random/issues/25))

IMPROVEMENTS:

* Add optional minimum character constraints to `random_string` ([#22](https://github.com/terraform-providers/terraform-provider-random/issues/22))

## 1.2.0 (April 03, 2018)

NEW FEATURES:

* `random_integer` and `random_id` are now importable. ([#20](https://github.com/terraform-providers/terraform-provider-random/issues/20))

## 1.1.0 (December 01, 2017)

NEW FEATURES:

* `random_integer` resource generates a single integer within a given range. ([#12](https://github.com/terraform-providers/terraform-provider-random/issues/12))

## 1.0.0 (September 15, 2017)

NEW FEATURES:

* `random_string` resource generates random strings of a given length consisting of letters, digits and symbols. ([#5](https://github.com/terraform-providers/terraform-provider-random/issues/5))

## 0.1.0 (June 21, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
