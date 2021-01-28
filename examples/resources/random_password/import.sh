# Random Password can be imported by specifying the value of the string alongside the resource attributes.
terraform import random_password.password "password keepers=nil,length=8,special=true,upper=true,lower=true,number=true,min_numeric=0,min_upper=0,min_lower=0,min_special=0,override_special=_%@"

# Note: `override_special` should be the last specified key if any

# Note: When importing keepers, a jsonString should be specified without spaces
terraform import random_password.test "234567 length=6,special=false,keepers=={\"bla\":\"dibla\",\"key\":\"value\"}"