// Subnet for CFCR
resource "azurerm_subnet" "cf-subnet" {
  name                 = "${var.env_id}-cf-sn"
  resource_group_name  = "${azurerm_resource_group.bosh.name}"
  virtual_network_name = "${azurerm_virtual_network.bosh.name}"
  address_prefix       = "${cidrsubnet(var.network_cidr, 4, 1)}"
}
