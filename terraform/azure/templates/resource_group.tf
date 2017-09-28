resource "azurerm_resource_group" "bosh" {
  name     = "${var.env_id}-bosh"
  location = "${var.location}"

  tags {
    environment = "${var.env_id}"
  }
}

resource "azurerm_public_ip" "bosh" {
  name                         = "${var.env_id}-bosh"
  location                     = "${var.location}"
  resource_group_name          = "${azurerm_resource_group.bosh.name}"
  public_ip_address_allocation = "static"

  tags {
    environment = "${var.env_id}"
  }
}
