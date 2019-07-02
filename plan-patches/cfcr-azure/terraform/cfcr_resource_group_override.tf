
resource "azurerm_resource_group" "cfcr" {
  name     = "${var.env_id}-cfcr"
  location = "${var.region}"

  tags {
    environment = "${var.env_id}"
  }
}
