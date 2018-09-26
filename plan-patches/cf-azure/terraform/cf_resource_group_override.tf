
resource "azurerm_resource_group" "cf" {
  name     = "${var.env_id}-cf"
  location = "${var.region}"

  tags {
    environment = "${var.env_id}"
  }
}
