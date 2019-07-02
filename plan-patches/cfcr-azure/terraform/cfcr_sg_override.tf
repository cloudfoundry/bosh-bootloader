
// Security Group For CFCR Nodes
resource "azurerm_network_security_group" "cfcr-master" {
  name                = "${var.env_id}-cfcr-master-sg"
  location            = "${var.region}"
  resource_group_name          = "${azurerm_resource_group.cfcr.name}"

  security_rule {
    name                       = "master"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "8443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}
