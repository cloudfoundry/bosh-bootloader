output "vnet_name" {
  value = "${azurerm_virtual_network.bosh.name}"
}

output "subnet_name" {
  value = "${azurerm_subnet.bosh.name}"
}

output "resource_group_name" {
  value = "${azurerm_resource_group.bosh.name}"
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

locals {
  director_name = "bosh-${var.env_id}"
  internal_cidr = "${var.internal_cidr}"
  internal_gw = "${cidrhost(local.internal_cidr, 1)}"
  jumpbox_internal_ip = "${cidrhost(local.internal_cidr, 5)}"
  director_internal_ip = "${cidrhost(local.internal_cidr, 6)}"
}

output "director_name" {
  value = "${local.director_name}"
}

output "internal_cidr" {
  value = "${local.internal_cidr}"
}

output "internal_gw" {
  value = "${local.internal_gw}"
}

output "jumpbox__internal_ip" {
  value = "${local.jumpbox_internal_ip}"
}

output "director__internal_ip" {
  value = "${local.director_internal_ip}"
}

