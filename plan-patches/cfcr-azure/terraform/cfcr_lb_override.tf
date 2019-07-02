resource "azurerm_public_ip" "cfcr-balancer-ip" {
  name                         = "${var.env_id}-cfcr-lb-ip"
  location                     = "${var.region}"
  resource_group_name          = "${azurerm_resource_group.cfcr.name}"
  public_ip_address_allocation = "static"
  sku                          = "Standard"
}

resource "azurerm_lb" "cfcr-balancer" {
  name                = "${var.env_id}-cfcr-lb"
  location            = "${var.region}"
  sku                 = "Standard"
  resource_group_name = "${azurerm_resource_group.cfcr.name}"

  frontend_ip_configuration {
    name                 = "${azurerm_public_ip.cfcr-balancer-ip.name}"
    public_ip_address_id = "${azurerm_public_ip.cfcr-balancer-ip.id}"
  }
}

resource "azurerm_lb_backend_address_pool" "cfcr-balancer-backend-pool" {
  name                = "${var.env_id}-cfcr-backend-pool"
  resource_group_name = "${azurerm_resource_group.cfcr.name}"
  loadbalancer_id     = "${azurerm_lb.cfcr-balancer.id}"
}

resource "azurerm_lb_probe" "api-health-probe" {
  name                = "${var.env_id}-api-health-probe"
  resource_group_name = "${azurerm_resource_group.cfcr.name}"
  loadbalancer_id     = "${azurerm_lb.cfcr-balancer.id}"
  protocol            = "Tcp"
  interval_in_seconds = 5
  number_of_probes    = 2
  port                = 8443
}

resource "azurerm_lb_rule" "cfcr-balancer-api-rule" {
  name                           = "${var.env_id}-api-rule"
  resource_group_name            = "${azurerm_resource_group.cfcr.name}"
  loadbalancer_id                = "${azurerm_lb.cfcr-balancer.id}"
  protocol                       = "Tcp"
  frontend_port                  = 8443
  backend_port                   = 8443
  frontend_ip_configuration_name = "${azurerm_public_ip.cfcr-balancer-ip.name}"
  probe_id                       = "${azurerm_lb_probe.api-health-probe.id}"
  backend_address_pool_id        = "${azurerm_lb_backend_address_pool.cfcr-balancer-backend-pool.id}"
}
