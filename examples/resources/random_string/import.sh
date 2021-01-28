# Strings can be imported by just specifying the value of the string:
terraform import random_string.test test

# Strings can be imported by just specifying the value of the string alongside the resource attributes.
terraform import random_string.test "test length=4"

# Strings can be imported by specifying the value of the string alongside the resource attributes.
terraform import random_string.test "test keepers=nil,length=4,special=true,upper=true,lower=true,number=true,min_numeric=0,min_upper=0,min_lower=0,min_special=0,override_special=_%@"

# Note: `override_special` should be the last specified key if any

# Note: When importing keepers, a jsonString should be specified without spaces
terreform import random_string.test "234567 length=6,special=false,keepers=={\"bla\":\"dibla\",\"key\":\"value\"}"