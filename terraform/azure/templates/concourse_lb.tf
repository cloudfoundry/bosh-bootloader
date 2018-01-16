resource "azurerm_subnet" "concourse-sn" {
  name                 = "${var.env_id}-concourse-sn"
  address_prefix       = "${cidrsubnet(var.network_cidr, 8, 1)}"
  resource_group_name  = "${azurerm_resource_group.bosh.name}"
  virtual_network_name = "${azurerm_virtual_network.bosh.name}"
}

resource "azurerm_public_ip" "ip" {
  name                         = "${var.env_id}-concourse-lb-ip"
  location                     = "${var.region}"
  resource_group_name          = "${azurerm_resource_group.bosh.name}"
  public_ip_address_allocation = "dynamic"
}

resource "azurerm_lb" "concourse" {
  name                = "${var.env_id}-concourse-lb"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  location            = "${var.region}"

  frontend_ip_configuration {
    name                 = "${var.env_id}-concourse-frontend-ip-configuration"
    public_ip_address_id = "${azurerm_public_ip.ip.id}"
  }
}

output "load_balancer" {
  value = "${azurerm_lb.concourse.name}"
}
