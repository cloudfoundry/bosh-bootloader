resource "azurerm_virtual_network" "bosh" {
  name                = "${var.env_id}-bosh-vn"
  address_space       = ["10.0.0.0/16"]
  location            = "${var.location}"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
}

resource "azurerm_subnet" "bosh" {
  name                 = "${var.env_id}-bosh-sn"
  address_prefix       = "10.0.0.0/16"
  resource_group_name  = "${azurerm_resource_group.bosh.name}"
  virtual_network_name = "${azurerm_virtual_network.bosh.name}"
}
