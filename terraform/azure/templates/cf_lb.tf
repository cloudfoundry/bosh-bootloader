variable "system_domain" {}

variable "pfx_cert_base64" {}

variable "pfx_password" {}

resource "azurerm_subnet" "cf-sn" {
  name                 = "${var.env_id}-cf-sn"
  address_prefix       = cidrsubnet(var.network_cidr, 8, 1)
  resource_group_name  = azurerm_resource_group.bosh.name
  virtual_network_name = azurerm_virtual_network.bosh.name
}

resource "azurerm_network_security_group" "cf" {
  name                = "${var.env_id}-cf"
  location            = var.region
  resource_group_name = azurerm_resource_group.bosh.name

  tags {
    environment = var.env_id
  }
}

resource "azurerm_network_security_rule" "cf-http" {
  name                        = "${var.env_id}-cf-http"
  priority                    = 201
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "80"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.bosh.name
  network_security_group_name = azurerm_network_security_group.cf.name
}

resource "azurerm_network_security_rule" "cf-https" {
  name                        = "${var.env_id}-cf-https"
  priority                    = 202
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "443"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.bosh.name
  network_security_group_name = azurerm_network_security_group.cf.name
}

resource "azurerm_network_security_rule" "cf-log" {
  name                        = "${var.env_id}-cf-log"
  priority                    = 203
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "4443"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.bosh.name
  network_security_group_name = azurerm_network_security_group.cf.name
}

resource "azurerm_public_ip" "cf" {
  name                         = "${var.env_id}-cf-lb-ip"
  location                     = var.region
  resource_group_name          = azurerm_resource_group.bosh.name
  public_ip_address_allocation = "dynamic"
}

resource "azurerm_application_gateway" "cf" {
  name                = "${var.env_id}-app-gateway"
  resource_group_name = azurerm_resource_group.bosh.name
  location            = var.region

  sku {
    name     = "Standard_Small"
    tier     = "Standard"
    capacity = 2
  }

  probe {
    name                = "health-probe"
    protocol            = "Http"
    path                = "/"
    host                = "api.${var.system_domain}"
    interval            = 30
    timeout             = 30
    unhealthy_threshold = 3
  }

  gateway_ip_configuration {
    name      = "${var.env_id}-cf-gateway-ip-configuration"
    subnet_id = "${azurerm_virtual_network.bosh.id}/subnets/${azurerm_subnet.cf-sn.name}"
  }

  frontend_port {
    name = "frontendporthttps"
    port = 443
  }

  frontend_port {
    name = "frontendporthttp"
    port = 80
  }

  frontend_port {
    name = "frontendportlogs"
    port = 4443
  }

  frontend_ip_configuration {
    name                 = "${var.env_id}-cf-frontend-ip-configuration"
    public_ip_address_id = azurerm_public_ip.cf.id
  }

  backend_address_pool {
    name = "${var.env_id}-cf-backend-address-pool"
  }

  backend_http_settings {
    name                  = "${azurerm_virtual_network.bosh.name}-be-htst"
    cookie_based_affinity = "Disabled"
    port                  = 80
    protocol              = "Http"
    request_timeout       = 10
    probe_name            = "health-probe"
  }

  ssl_certificate {
    name     = "ssl-cert"
    data     = var.pfx_cert_base64
    password = var.pfx_password
  }

  http_listener {
    name                           = "${azurerm_virtual_network.bosh.name}-http-lstn"
    frontend_ip_configuration_name = "${var.env_id}-cf-frontend-ip-configuration"
    frontend_port_name             = "frontendporthttp"
    protocol                       = "Http"
  }

  http_listener {
    name                           = "${azurerm_virtual_network.bosh.name}-https-lstn"
    frontend_ip_configuration_name = "${var.env_id}-cf-frontend-ip-configuration"
    frontend_port_name             = "frontendporthttps"
    protocol                       = "Https"
    ssl_certificate_name           = "ssl-cert"
  }

  http_listener {
    name                           = "${azurerm_virtual_network.bosh.name}-logs-lstn"
    frontend_ip_configuration_name = "${var.env_id}-cf-frontend-ip-configuration"
    frontend_port_name             = "frontendportlogs"
    protocol                       = "Https"
    ssl_certificate_name           = "ssl-cert"
  }

  request_routing_rule {
    name                       = "${azurerm_virtual_network.bosh.name}-http-rule"
    rule_type                  = "Basic"
    http_listener_name         = "${azurerm_virtual_network.bosh.name}-http-lstn"
    backend_address_pool_name  = "${var.env_id}-cf-backend-address-pool"
    backend_http_settings_name = "${azurerm_virtual_network.bosh.name}-be-htst"
  }

  request_routing_rule {
    name                       = "${azurerm_virtual_network.bosh.name}-https-rule"
    rule_type                  = "Basic"
    http_listener_name         = "${azurerm_virtual_network.bosh.name}-https-lstn"
    backend_address_pool_name  = "${var.env_id}-cf-backend-address-pool"
    backend_http_settings_name = "${azurerm_virtual_network.bosh.name}-be-htst"
  }

  request_routing_rule {
    name                       = "${azurerm_virtual_network.bosh.name}-logs-rule"
    rule_type                  = "Basic"
    http_listener_name         = "${azurerm_virtual_network.bosh.name}-logs-lstn"
    backend_address_pool_name  = "${var.env_id}-cf-backend-address-pool"
    backend_http_settings_name = "${azurerm_virtual_network.bosh.name}-be-htst"
  }
}

output "cf_app_gateway_name" {
  value = azurerm_application_gateway.cf.name
}

output "cf_security_group" {
  value = azurerm_network_security_group.cf.name
}
