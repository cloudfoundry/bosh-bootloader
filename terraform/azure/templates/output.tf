output "vnet_name" {
  value = "${var.vnet_name == "" ? join("", azurerm_virtual_network.bosh.*.name) : var.vnet_name}"
}

output "subnet_name" {
  value = "${var.subnet_name == "" ? join("", azurerm_subnet.bosh.*.name) : var.subnet_name}"
}

output "resource_group_name" {
  value = "${var.resource_group_name == "" ? join("", azurerm_resource_group.bosh.*.name) : var.resource_group_name}"
}

output "vnet_resource_group_name" {
  value = "${var.vnet_resource_group_name}"
}

output "storage_account_name" {
  value = "${azurerm_storage_account.bosh.name}"
}

output "default_security_group" {
  value = "${azurerm_network_security_group.bosh.name}"
}

output "external_ip" {
  value = "${azurerm_public_ip.bosh.ip_address}"
}

output "director_address" {
  value = "https://${azurerm_public_ip.bosh.ip_address}:25555"
}

output "private_key" {
  value     = "${tls_private_key.bosh_vms.private_key_pem}"
  sensitive = true
}

output "public_key" {
  value     = "${tls_private_key.bosh_vms.public_key_openssh}"
  sensitive = false
}

output "jumpbox_url" {
  value = "${azurerm_public_ip.bosh.ip_address}:22"
}

output "network_cidr" {
  value = "${var.network_cidr}"
}

output "director_name" {
  value = "bosh-${var.env_id}"
}

output "internal_cidr" {
  value = "${var.internal_cidr}"
}

output "subnet_cidr" {
  value = "${cidrsubnet(var.network_cidr, 8, 0)}"
}

output "internal_gw" {
  value = "${cidrhost(var.internal_cidr, 1)}"
}

output "jumpbox__internal_ip" {
  value = "${cidrhost(var.internal_cidr, 5)}"
}

output "director__internal_ip" {
  value = "${cidrhost(var.internal_cidr, 6)}"
}
