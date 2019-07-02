variable "cfcr_internal_cidr" {
  default = "10.0.16.0/20"
}
output "cfcr_master_loadbalancer_name" {
  value = "${var.env_id}-cfcr-lb"
}

output "api-hostname" {
  value = "${azurerm_public_ip.cfcr-balancer-ip.ip_address}"
}

output "vnet_resource_group_name" {
  value = "${azurerm_resource_group.bosh.name}"
}

// CFCR Subnet
output "cfcr_subnet" {
  value = "${azurerm_subnet.cfcr-subnet.name}"
}

output "cfcr_subnet_cidr" {
  value = "${cidrsubnet(var.network_cidr, 4, 1)}"
}

output "cfcr_internal_gw" {
  value = "${cidrhost(var.cfcr_internal_cidr, 1)}"
}

output "cfcr_master_security_group" {
  value = "${azurerm_network_security_group.cfcr-master.name}"
}

// Cloud Providers
output "azure_cloud_name" {
    value = "AzurePublicCloud"
}

output "subscription_id" {
    value = "${var.subscription_id}"
}

output "tenant_id" {
    value = "${var.tenant_id}"
}

output "client_id" {
    value = "${var.client_id}"
}

output "client_secret" {
    value = "${var.client_secret}"
}

output "cfcr_resource_group_name" {
    value = "${azurerm_resource_group.cfcr.name}"
}

output "cfcr_vnet_resource_group_name" {
    value = "${azurerm_resource_group.bosh.name}"
}

output "cfcr_vnet_name" {
    value = "${azurerm_virtual_network.bosh.name}"
}

output "cfcr_subnet_name" {
    value = "${azurerm_subnet.cfcr-subnet.name}"
}

output "primary_availability_set" {
    value = "bosh-${var.env_id}-azurecfcr-worker"
}
