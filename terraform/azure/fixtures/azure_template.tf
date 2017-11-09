variable "env_id" {}

variable "location" {}

variable "simple_env_id" {}

variable "subscription_id" {}

variable "tenant_id" {}

variable "client_id" {}

variable "client_secret" {}

variable "network_cidr" {
  default = "10.0.0.0/16"
}

variable "internal_cidr" {
  default = "10.0.0.0/24"
}

provider "azurerm" {
  subscription_id  = "${var.subscription_id}"
  tenant_id        = "${var.tenant_id}"
  client_id        = "${var.client_id}"
  client_secret    = "${var.client_secret}"
}

resource "azurerm_resource_group" "bosh" {
  name     = "${var.env_id}-bosh"
  location = "${var.location}"

  tags {
    environment = "${var.env_id}"
  }
}

resource "azurerm_public_ip" "bosh" {
  name                         = "${var.env_id}-bosh"
  location                     = "${var.location}"
  resource_group_name          = "${azurerm_resource_group.bosh.name}"
  public_ip_address_allocation = "static"

  tags {
    environment = "${var.env_id}"
  }
}

resource "azurerm_virtual_network" "bosh" {
  name                = "${var.env_id}-bosh-vn"
  address_space       = ["${var.network_cidr}"]
  location            = "${var.location}"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
}

resource "azurerm_subnet" "bosh" {
  name                 = "${var.env_id}-bosh-sn"
  address_prefix       = "${var.network_cidr}"
  resource_group_name  = "${azurerm_resource_group.bosh.name}"
  virtual_network_name = "${azurerm_virtual_network.bosh.name}"
}

resource "azurerm_storage_account" "bosh" {
  name                = "${var.simple_env_id}"
  resource_group_name = "${azurerm_resource_group.bosh.name}"

  location     = "westus"
  account_tier = "Standard"
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

resource "azurerm_network_security_group" "bosh" {
  name                = "${var.env_id}-bosh"
  location            = "${var.location}"
  resource_group_name = "${azurerm_resource_group.bosh.name}"

  tags {
    environment = "${var.env_id}"
  }
}

resource "azurerm_network_security_group" "cf" {
  name                = "${var.env_id}-cf"
  location            = "${var.location}"
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

resource "azurerm_network_security_rule" "credhub" {
  name                       = "${var.env_id}-credhub"
  priority                   = 204
  direction                  = "Inbound"
  access                     = "Allow"
  protocol                   = "Tcp"
  source_port_range          = "*"
  destination_port_range     = "8844"
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

output "bosh_network_name" {
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

output "bosh_vms_private_key" {
  value = "${tls_private_key.bosh_vms.private_key_pem}"
  sensitive = true
}

output "bosh_vms_public_key" {
  value = "${tls_private_key.bosh_vms.public_key_openssh}"
  sensitive = false
}

output "jumpbox_url" {
	value = "${azurerm_public_ip.bosh.ip_address}:22"
}

output "network_cidr" {
  value = "${var.network_cidr}"
}

output "internal_cidr" {
  value = "${var.internal_cidr}"
}

resource "tls_private_key" "bosh_vms" {
  algorithm = "RSA"
  rsa_bits = 4096
}
