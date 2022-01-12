resource "azurerm_network_security_group" "bosh" {
  name                = "${var.env_id}-bosh"
  location            = var.region
  resource_group_name = azurerm_resource_group.bosh.name

  tags {
    environment = var.env_id
  }
}

resource "azurerm_network_security_rule" "ssh" {
  name                        = "${var.env_id}-ssh"
  priority                    = 200
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "22"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.bosh.name
  network_security_group_name = azurerm_network_security_group.bosh.name
}

resource "azurerm_network_security_rule" "bosh-agent" {
  name                        = "${var.env_id}-bosh-agent"
  priority                    = 201
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "6868"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.bosh.name
  network_security_group_name = azurerm_network_security_group.bosh.name
}

resource "azurerm_network_security_rule" "bosh-director" {
  name                        = "${var.env_id}-bosh-director"
  priority                    = 202
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "25555"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.bosh.name
  network_security_group_name = azurerm_network_security_group.bosh.name
}

resource "azurerm_network_security_rule" "dns" {
  name                        = "${var.env_id}-dns"
  priority                    = 203
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "53"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.bosh.name
  network_security_group_name = azurerm_network_security_group.bosh.name
}

resource "azurerm_network_security_rule" "credhub" {
  name                        = "${var.env_id}-credhub"
  priority                    = 204
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "8844"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.bosh.name
  network_security_group_name = azurerm_network_security_group.bosh.name
}
