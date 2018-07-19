variable "existing_resource_group" {}

data "azurerm_resource_group" "bosh" {
  name = "${var.existing_resource_group}"
}

resource "azurerm_resource_group" "bosh" {
  count = 0
}

resource "azurerm_public_ip" "bosh" {
  resource_group_name = "${data.azurerm_resource_group.bosh.name}"
}

resource "azurerm_virtual_network" "bosh" {
  resource_group_name = "${data.azurerm_resource_group.bosh.name}"
}

resource "azurerm_subnet" "bosh" {
  resource_group_name = "${data.azurerm_resource_group.bosh.name}"
}

resource "azurerm_storage_account" "bosh" {
  resource_group_name = "${data.azurerm_resource_group.bosh.name}"
}

resource "azurerm_storage_container" "bosh" {
  resource_group_name = "${data.azurerm_resource_group.bosh.name}"
}

resource "azurerm_storage_container" "stemcell" {
  resource_group_name = "${data.azurerm_resource_group.bosh.name}"
}

resource "azurerm_network_security_group" "bosh" {
  resource_group_name = "${data.azurerm_resource_group.bosh.name}"
}

resource "azurerm_network_security_rule" "ssh" {
  resource_group_name = "${data.azurerm_resource_group.bosh.name}"
}

resource "azurerm_network_security_rule" "bosh-agent" {
  resource_group_name = "${data.azurerm_resource_group.bosh.name}"
}

resource "azurerm_network_security_rule" "bosh-director" {
  resource_group_name = "${data.azurerm_resource_group.bosh.name}"
}

resource "azurerm_network_security_rule" "dns" {
  resource_group_name = "${data.azurerm_resource_group.bosh.name}"
}

resource "azurerm_network_security_rule" "credhub" {
  resource_group_name = "${data.azurerm_resource_group.bosh.name}"
}

output "resource_group_name" {
  value = "${data.azurerm_resource_group.bosh.name}"
}
