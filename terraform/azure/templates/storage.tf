resource "azurerm_storage_account" "bosh" {
  name                = "${var.simple_env_id}"
  resource_group_name = "${azurerm_resource_group.bosh.name}"

  location                 = "westus"
  account_tier             = "Standard"
  account_replication_type = "GRS"

  tags {
    environment = "${var.env_id}"
  }
}

resource "azurerm_storage_container" "bosh" {
  name                  = "bosh"
  resource_group_name   = "${azurerm_resource_group.bosh.name}"
  storage_account_name  = "${azurerm_storage_account.bosh.name}"
  container_access_type = "private"
}

resource "azurerm_storage_container" "stemcell" {
  name                  = "stemcell"
  resource_group_name   = "${azurerm_resource_group.bosh.name}"
  storage_account_name  = "${azurerm_storage_account.bosh.name}"
  container_access_type = "blob"
}
