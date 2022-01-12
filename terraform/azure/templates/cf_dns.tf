data "azurerm_public_ip" "cf-lb" {
  name                = "${var.env_id}-cf-lb-ip"
  resource_group_name = azurerm_resource_group.bosh.name
  depends_on          = ["azurerm_application_gateway.cf"]
}

resource "azurerm_dns_zone" "cf" {
  name                = var.system_domain
  resource_group_name = azurerm_resource_group.bosh.name

  tags {
    environment = var.env_id
  }
}

resource "azurerm_dns_a_record" "cf" {
  name                = "*"
  zone_name           = azurerm_dns_zone.cf.name
  resource_group_name = azurerm_resource_group.bosh.name
  ttl                 = "300"
  records             = ["${data.azurerm_public_ip.cf-lb.ip_address}"]
}

resource "azurerm_dns_a_record" "bosh" {
  name                = "bosh"
  zone_name           = azurerm_dns_zone.cf.name
  resource_group_name = azurerm_resource_group.bosh.name
  ttl                 = "300"
  records             = ["${azurerm_public_ip.bosh.ip_address}"]
}
