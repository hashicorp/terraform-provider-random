# The following example shows how to generate a unique name for an Azure Resource Group.

resource "random_uuid7" "test" {
}

resource "azurerm_resource_group" "test" {
  name     = "${random_uuid7.test.result}-rg"
  location = "Central US"
}
