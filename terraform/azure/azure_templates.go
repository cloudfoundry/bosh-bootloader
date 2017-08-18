package azure

const VarsTemplate = `
variable "env_id" {
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

const ResourceGroupTemplate = `
resource "azurerm_resource_group" "test" {
  name     = "${var.env_id}-test"
  location = "West US"

  tags {
    environment = "${var.env_id}-test"
  }
}`

const NetworkTemplate = `# Create a virtual network in the web_servers resource group
resource "azurerm_virtual_network" "network" {
  name                = "boshnet"
  address_space       = ["10.0.0.0/16"]
  location            = "West US"
  resource_group_name = "${var.env_id}-test"

  subnet {
    name           = "subnet1"
    address_prefix = "10.0.1.0/24"
  }
}`

const StorageTemplate = `# https://www.terraform.io/docs/providers/azurerm/r/storage_account.html
resource "azurerm_storage_account" "storage" {
  name                = "boshstore"
  resource_group_name = "${var.env_id}-test"

  location     = "westus"
  account_type = "Standard_GRS"

  tags {
    environment = "${var.env_id}-test"
  }
}`

const NetworkSecurityGroupTemplate = ` # https://www.terraform.io/docs/providers/azurerm/r/network_security_group.html
resource "azurerm_network_security_group" "security_group" {
  name                = "nsg-bosh"
  location            = "West US"
  resource_group_name = "${var.env_id}-test"

  security_rule {
    name                       = "nsg-bosh"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  tags {
    environment = "Production"
  }
}`

const OutputTemplate = `
output "bosh_network_name" {
    value = "${azurerm_virtual_network.network.name}"
}

output "bosh_subnet_name" {
    value = "${azurerm_virtual_network.network.subnet.name}"
}

output "bosh_resource_group_name" {
    value = "${azurerm_resource_group.test.name}"
}

output "bosh_storage_account_name" {
    value = "${azurerm_storage_account.storage.name}"
}

output "bosh_default_security_group" {
    value = "${azurerm_network_security_group.security_group.name}"
}
`
