resource "azurerm_virtual_network" "bosh" {
  name                = "${var.env_id}-bosh-vn"
  address_space       = ["${var.network_cidr}"]
  location            = "${var.region}"
  resource_group_name = "${var.resource_group_name == "" ? azurerm_resource_group.bosh.*.name[0] : var.resource_group_name}"
  count               = "${var.vnet_name == "" ? 1 : 0 }"
}


resource "azurerm_subnet" "bosh" {
  name                 = "${var.env_id}-bosh-sn"
  address_prefix       = "${cidrsubnet(var.network_cidr, 8, 0)}"
  resource_group_name  = "${var.resource_group_name == "" ? azurerm_resource_group.bosh.*.name[0] : var.resource_group_name}"
  virtual_network_name = "${var.vnet_name == "" ? azurerm_virtual_network.bosh.*.name[0] : var.vnet_name}"
}

