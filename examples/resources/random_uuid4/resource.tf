# The following example shows how to generate a unique name for an Azure Resource Group.

resource "random_uuid4" "test" {
}

resource "azurerm_resource_group" "test" {
  name     = "${random_uuid4.test.result}-rg"
  location = "Central US"
}
