package azure

const VarsTemplate = `variable "env_id" {
	type = "string"
}

variable "simple_env_id" {
	type = "string"
}

variable "subscription_id" {
	type = "string"
}

variable "tenant_id" {
	type = "string"
}

variable "client_id" {
	type = "string"
}

variable "client_secret" {
	type = "string"
}

provider "azurerm" {
  subscription_id  = "${var.subscription_id}"
  tenant_id        = "${var.tenant_id}"
  client_id        = "${var.client_id}"
  client_secret    = "${var.client_secret}"
}
`

const ResourceGroupTemplate = `resource "azurerm_resource_group" "bosh" {
  name     = "${var.env_id}-bosh"
  location = "West US"

  tags {
    environment = "${var.env_id}"
  }
}

resource "azurerm_public_ip" "bosh" {
  name                         = "${var.env_id}-bosh"
  location                     = "West US"
  resource_group_name          = "${azurerm_resource_group.bosh.name}"
  public_ip_address_allocation = "static"

  tags {
    environment = "${var.env_id}"
  }
}
`

const NetworkTemplate = `resource "azurerm_virtual_network" "bosh" {
  name                = "${var.env_id}-bosh-vn"
  address_space       = ["10.0.0.0/16"]
  location            = "West US"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
}

resource "azurerm_subnet" "bosh" {
  name                 = "${var.env_id}-bosh-sn"
  address_prefix       = "10.0.0.0/16"
  resource_group_name  = "${azurerm_resource_group.bosh.name}"
  virtual_network_name = "${azurerm_virtual_network.bosh.name}"
}
`

const StorageTemplate = `resource "azurerm_storage_account" "bosh" {
  name                = "${var.simple_env_id}"
  resource_group_name = "${azurerm_resource_group.bosh.name}"

  location     = "westus"
  account_type = "Standard_GRS"

  tags {
    environment = "${var.env_id}"
  }
}

resource "azurerm_storage_container" "bosh" {
  name                  = "${var.env_id}-bosh"
  resource_group_name   = "${azurerm_resource_group.bosh.name}"
  storage_account_name  = "${azurerm_storage_account.bosh.name}"
  container_access_type = "private"
}

resource "azurerm_storage_container" "stemcell" {
  name                  = "${var.env_id}-stemcell"
  resource_group_name   = "${azurerm_resource_group.bosh.name}"
  storage_account_name  = "${azurerm_storage_account.bosh.name}"
  container_access_type = "blob"
}
`

const NetworkSecurityGroupTemplate = `resource "azurerm_network_security_group" "bosh" {
  name                = "${var.env_id}-bosh"
  location            = "West US"
  resource_group_name = "${azurerm_resource_group.bosh.name}"

  tags {
    environment = "${var.env_id}"
  }
}

resource "azurerm_network_security_group" "cf" {
  name                = "${var.env_id}-cf"
  location            = "West US"
  resource_group_name = "${azurerm_resource_group.bosh.name}"

  tags {
    environment = "${var.env_id}"
  }
}

resource "azurerm_network_security_rule" "ssh" {
  name                       = "${var.env_id}-ssh"
  priority                   = 200
  direction                  = "Inbound"
  access                     = "Allow"
  protocol                   = "Tcp"
  source_port_range          = "*"
  destination_port_range     = "22"
  source_address_prefix      = "*"
  destination_address_prefix = "*"
  resource_group_name         = "${azurerm_resource_group.bosh.name}"
  network_security_group_name = "${azurerm_network_security_group.bosh.name}"
}

resource "azurerm_network_security_rule" "bosh-agent" {
  name                       = "${var.env_id}-bosh-agent"
  priority                   = 201
  direction                  = "Inbound"
  access                     = "Allow"
  protocol                   = "Tcp"
  source_port_range          = "*"
  destination_port_range     = "6868"
  source_address_prefix      = "*"
  destination_address_prefix = "*"
  resource_group_name         = "${azurerm_resource_group.bosh.name}"
  network_security_group_name = "${azurerm_network_security_group.bosh.name}"
}

resource "azurerm_network_security_rule" "bosh-director" {
  name                       = "${var.env_id}-bosh-director"
  priority                   = 202
  direction                  = "Inbound"
  access                     = "Allow"
  protocol                   = "Tcp"
  source_port_range          = "*"
  destination_port_range     = "25555"
  source_address_prefix      = "*"
  destination_address_prefix = "*"
  resource_group_name         = "${azurerm_resource_group.bosh.name}"
  network_security_group_name = "${azurerm_network_security_group.bosh.name}"
}

resource "azurerm_network_security_rule" "dns" {
  name                       = "${var.env_id}-dns"
  priority                   = 203
  direction                  = "Inbound"
  access                     = "Allow"
  protocol                   = "*"
  source_port_range          = "*"
  destination_port_range     = "53"
  source_address_prefix      = "*"
  destination_address_prefix = "*"
  resource_group_name         = "${azurerm_resource_group.bosh.name}"
  network_security_group_name = "${azurerm_network_security_group.bosh.name}"
}

resource "azurerm_network_security_rule" "cf-https" {
  name                       = "${var.env_id}-dns"
  priority                   = 201
  direction                  = "Inbound"
  access                     = "Allow"
  protocol                   = "Tcp"
  source_port_range          = "*"
  destination_port_range     = "443"
  source_address_prefix      = "*"
  destination_address_prefix = "*"
  resource_group_name         = "${azurerm_resource_group.bosh.name}"
  network_security_group_name = "${azurerm_network_security_group.cf.name}"
}

resource "azurerm_network_security_rule" "cf-log" {
  name                       = "${var.env_id}-cf-log"
  priority                   = 202
  direction                  = "Inbound"
  access                     = "Allow"
  protocol                   = "Tcp"
  source_port_range          = "*"
  destination_port_range     = "4443"
  source_address_prefix      = "*"
  destination_address_prefix = "*"
  resource_group_name         = "${azurerm_resource_group.bosh.name}"
  network_security_group_name = "${azurerm_network_security_group.cf.name}"
}
`

const OutputTemplate = `output "bosh_network_name" {
    value = "${azurerm_virtual_network.bosh.name}"
}

output "bosh_subnet_name" {
    value = "${azurerm_subnet.bosh.name}"
}

output "bosh_resource_group_name" {
    value = "${azurerm_resource_group.bosh.name}"
}

output "bosh_storage_account_name" {
    value = "${azurerm_storage_account.bosh.name}"
}

output "bosh_default_security_group" {
    value = "${azurerm_network_security_group.bosh.name}"
}

output "external_ip" {
    value = "${azurerm_public_ip.bosh.ip_address}"
}

output "director_address" {
	value = "https://${azurerm_public_ip.bosh.ip_address}:25555"
}
`
