resource "azurerm_public_ip" "concourse" {
  name                         = "${var.env_id}-concourse-lb"
  location                     = "${var.region}"
  resource_group_name          = "${azurerm_resource_group.bosh.name}"
  public_ip_address_allocation = "static"
  sku                          = "Standard"

  tags {
    environment = "${var.env_id}"
  }
}

resource "azurerm_lb" "concourse" {
  name                = "${var.env_id}-concourse-lb"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  location            = "${var.region}"
  sku                 = "Standard"

  frontend_ip_configuration {
    name                 = "${var.env_id}-concourse-frontend-ip-configuration"
    public_ip_address_id = "${azurerm_public_ip.concourse.id}"
  }
}

resource "azurerm_lb_rule" "concourse-https" {
  name                = "${var.env_id}-concourse-https"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  loadbalancer_id     = "${azurerm_lb.concourse.id}"

  frontend_ip_configuration_name = "${var.env_id}-concourse-frontend-ip-configuration"
  protocol                       = "TCP"
  frontend_port                  = 443
  backend_port                   = 443

  backend_address_pool_id = "${azurerm_lb_backend_address_pool.concourse.id}"
  probe_id                = "${azurerm_lb_probe.concourse-https.id}"
}

resource "azurerm_lb_probe" "concourse-https" {
  name                = "${var.env_id}-concourse-https"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  loadbalancer_id     = "${azurerm_lb.concourse.id}"
  protocol            = "TCP"
  port                = 443
}

resource "azurerm_lb_rule" "concourse-http" {
  name                = "${var.env_id}-concourse-http"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  loadbalancer_id     = "${azurerm_lb.concourse.id}"

  frontend_ip_configuration_name = "${var.env_id}-concourse-frontend-ip-configuration"
  protocol                       = "TCP"
  frontend_port                  = 80
  backend_port                   = 80

  backend_address_pool_id = "${azurerm_lb_backend_address_pool.concourse.id}"
  probe_id                = "${azurerm_lb_probe.concourse-http.id}"
}

resource "azurerm_lb_probe" "concourse-http" {
  name                = "${var.env_id}-concourse-http"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  loadbalancer_id     = "${azurerm_lb.concourse.id}"
  protocol            = "TCP"
  port                = 80
}

resource "azurerm_lb_rule" "concourse-uaa" {
  name                = "${var.env_id}-concourse-uaa"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  loadbalancer_id     = "${azurerm_lb.concourse.id}"

  frontend_ip_configuration_name = "${var.env_id}-concourse-frontend-ip-configuration"
  protocol                       = "TCP"
  frontend_port                  = 8443
  backend_port                   = 8443

  backend_address_pool_id = "${azurerm_lb_backend_address_pool.concourse.id}"
  probe_id                = "${azurerm_lb_probe.concourse-uaa.id}"
}

resource "azurerm_lb_probe" "concourse-uaa" {
  name                = "${var.env_id}-concourse-uaa"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  loadbalancer_id     = "${azurerm_lb.concourse.id}"
  protocol            = "TCP"
  port                = 8443
}

resource "azurerm_lb_rule" "concourse-credhub" {
  name                = "${var.env_id}-concourse-credhub"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  loadbalancer_id     = "${azurerm_lb.concourse.id}"

  frontend_ip_configuration_name = "${var.env_id}-concourse-frontend-ip-configuration"
  protocol                       = "TCP"
  frontend_port                  = 8844
  backend_port                   = 8844

  backend_address_pool_id = "${azurerm_lb_backend_address_pool.concourse.id}"
  probe_id                = "${azurerm_lb_probe.concourse-credhub.id}"
}

resource "azurerm_lb_probe" "concourse-credhub" {
  name                = "${var.env_id}-concourse-credhub"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  loadbalancer_id     = "${azurerm_lb.concourse.id}"
  protocol            = "TCP"
  port                = 8844
}

resource "azurerm_network_security_rule" "concourse-http" {
  name                        = "${var.env_id}-concourse-http"
  priority                    = 209
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "80"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurerm_resource_group.bosh.name}"
  network_security_group_name = "${azurerm_network_security_group.bosh.name}"
}

resource "azurerm_network_security_rule" "concourse-https" {
  name                        = "${var.env_id}-concourse-https"
  priority                    = 208
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "443"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurerm_resource_group.bosh.name}"
  network_security_group_name = "${azurerm_network_security_group.bosh.name}"
}

resource "azurerm_network_security_rule" "concourse-credhub" {
  name                        = "${var.env_id}-uaa"
  priority                    = 207
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "8844"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurerm_resource_group.bosh.name}"
  network_security_group_name = "${azurerm_network_security_group.bosh.name}"
}

resource "azurerm_network_security_rule" "uaa" {
  name                        = "${var.env_id}-uaa"
  priority                    = 206
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "8443"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurerm_resource_group.bosh.name}"
  network_security_group_name = "${azurerm_network_security_group.bosh.name}"
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
