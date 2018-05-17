resource "azurerm_resource_group" "bosh" {
  name     = "${var.resource_group_name}"
  location = "${var.region}"

  tags {
    environment = "${var.env_id}"
  }
}

resource "azurerm_public_ip" "bosh" {
  name                         = "${var.env_id}-bosh"
  location                     = "${var.region}"
  resource_group_name          = "${var.resource_group_name}"
  public_ip_address_allocation = "static"

  tags {
    environment = "${var.env_id}"
  }
}
