// Subnet for CFCR
resource "azurerm_subnet" "cfcr-subnet" {
  name                 = "${var.env_id}-cfcr-sn"
  resource_group_name  = "${azurerm_resource_group.bosh.name}"
  virtual_network_name = "${azurerm_virtual_network.bosh.name}"
  address_prefix       = "${cidrsubnet(var.network_cidr, 4, 1)}"
  network_security_group_id = "${azurerm_network_security_group.cfcr-master.id}"
}
