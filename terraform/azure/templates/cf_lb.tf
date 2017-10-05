# Create a application gateway in the web_servers resource group
# resource "azurerm_virtual_network" "vnet" {
#   name                = "my-vnet-12345"
#   resource_group_name = "${azurerm_resource_group.bosh.name}"
#   address_space       = ["10.254.0.0/16"]
#   location            = "${azurerm_resource_group.rg.location}"
# }

resource "azurerm_subnet" "sub1" {
  name                 = "subnet-1"
  resource_group_name  = "${azurerm_resource_group.bosh.name}"
  virtual_network_name = "${azurerm_virtual_network.bosh.name}"
  address_prefix       = "10.254.0.0/24"
}

resource "azurerm_subnet" "sub2" {
  name                 = "subnet-2"
  resource_group_name  = "${azurerm_resource_group.bosh.name}"
  virtual_network_name = "${azurerm_virtual_network.bosh.name}"
  address_prefix       = "10.254.2.0/24"
}

resource "azurerm_public_ip" "lb" {
  name                         = "${var.env_id}-lb"
  location                     = "${azurerm_resource_group.rg.location}"
  resource_group_name          = "${azurerm_resource_group.bosh.name}"
  public_ip_address_allocation = "dynamic"
}

# Create an application gateway
resource "azurerm_application_gateway" "network" {
  name                = "{var.env_id}-app-gateway"
  resource_group_name = "${azurerm_resource_group.bosh.name}"
  location            = "West US"
 
  sku {
    name           = "Standard_Small"
    tier           = "Standard"
    capacity       = 2
  }
 
  gateway_ip_configuration {
      name         = "{var.env_id}-gateway-ip-configuration"
      subnet_id    = "${azurerm_virtual_network.vnet.id}/subnets/${azurerm_subnet.sub1.name}"
  }
 
  frontend_port {
      name         = "{var.env_id}-frontend-port"
      port         = 80
  }
 
  frontend_ip_configuration {
      name         = "{var.env_id}-frontend-ip-configuration"  
      public_ip_address_id = "${azurerm_public_ip.lb.id}"
  }

  backend_address_pool {
      name = "my-backend-address-pool"
  }
 
  backend_http_settings {
      name                  = "${azurerm_virtual_network.vnet.name}-be-htst"
      cookie_based_affinity = "Disabled"
      port                  = 80
      protocol              = "Http"
     request_timeout        = 1
  }
 
  http_listener {
        name                                  = "${azurerm_virtual_network.vnet.name}-httplstn"
        frontend_ip_configuration_name        = "my-frontend-ip-configuration"
        frontend_port_name                    = "my-frontend-port"
        protocol                              = "Http"
  }
 
  request_routing_rule {
          name                       = "${azurerm_virtual_network.vnet.name}-rqrt"
          rule_type                  = "Basic"
          http_listener_name         = "${azurerm_virtual_network.vnet.name}-httplstn"
          backend_address_pool_name  = "my-backend-address-pool"
          backend_http_settings_name = "${azurerm_virtual_network.vnet.name}-be-htst"
  }
}