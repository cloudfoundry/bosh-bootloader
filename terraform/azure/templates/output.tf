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

output "internal_cidr" {
  value = "${var.internal_cidr}"
}
