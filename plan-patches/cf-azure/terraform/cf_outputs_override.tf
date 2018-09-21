output "vnet_name" {
  value = "${azurerm_virtual_network.bosh.name}"
}

output "cf_internal_gw" {
  value = "${cidrhost(cidrsubnet(var.network_cidr, 4, 1), 1)}"
}

output "cf_subnet_cidr" {
  value = "${cidrsubnet(var.network_cidr, 4, 1)}"
}

output "cf_subnet" {
  value = "${azurerm_subnet.cf-subnet.name}"
}

output "cf_loadbalancer_name" {
  value = "${var.env_id}-cf-lb"
}

output "cf_balancer_pub_ip" {
  value = "${azurerm_public_ip.cf-balancer-ip.ip_address}"
}

output "cf_security_group" {
  value = "${azurerm_network_security_group.cf.name}"
}

output "cf_resource_group_name" {
  value = "${azurerm_resource_group.cf.name}"
}

output "network_name" {
  value = "cf"
}