resource "azurerm_resource_group" "bosh" {
  name     = "${var.env_id}-bosh"
  location = "${var.region}"
  count    = "${var.resource_group_name == "" ? 1 : 0 }"

  tags {
    environment = "${var.env_id}"
  }
}

resource "azurerm_public_ip" "bosh" {
  name                         = "${var.env_id}-bosh"
  location                     = "${var.region}"
  resource_group_name          = "${var.resource_group_name == "" ? azurerm_resource_group.bosh.*.name[0] : var.resource_group_name}"
  public_ip_address_allocation = "static"

  tags {
    environment = "${var.env_id}"
  }
}
