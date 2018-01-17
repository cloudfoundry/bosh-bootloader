resource "azurerm_public_ip" "concourse" {
  name                         = "${var.env_id}-concourse-lb"
  location                     = "${var.region}"
  resource_group_name          = "${azurerm_resource_group.bosh.name}"
  public_ip_address_allocation = "static"
}

resource "azurerm_lb" "concourse" {
  name                = "${var.env_id}-concourse-lb"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  location            = "${var.region}"

  frontend_ip_configuration {
    name                 = "${var.env_id}-concourse-frontend-ip-configuration"
    public_ip_address_id = "${azurerm_public_ip.concourse.id}"
  }
}

resource "azurerm_lb_backend_address_pool" "concourse" {
  name                = "${var.env_id}-concourse-backend-pool"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  loadbalancer_id     = "${azurerm_lb.concourse.id}"
}

output "concourse_lb_name" {
  value = "${azurerm_lb.concourse.name}"
}

output "concourse_lb_ip" {
  value = "${azurerm_public_ip.concourse.ip_address}"
}
