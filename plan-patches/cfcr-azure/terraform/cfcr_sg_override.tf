
// Security Group For CFCR Master Nodes
resource "azurerm_network_security_group" "cfcr-master" {
  name                = "${var.env_id}-cfcr-master-sg"
  location            = "${var.region}"
  resource_group_name          = "${azurerm_resource_group.bosh.name}"

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

// Security Group For CFCR Worker Nodes
resource "azurerm_network_security_group" "cfcr-worker" {
  name                = "${var.env_id}-cfcr-worker-sg"
  location            = "${var.region}"
  # TODO: after the new cpi release, switch to cfcr resource group.
  resource_group_name          = "${azurerm_resource_group.bosh.name}"
}