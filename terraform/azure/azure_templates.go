package azure

const VarsTemplate = `
variable "env_id" {
	type = "string"
}

variable "subscription_id" {
	type = "string"
}

variable "tenant_id" {
	type = "string"
}

variable "client_id" {
	type = "string"
}

variable "client_secret" {
	type = "string"
}

provider "azurerm" {
  subscription_id="${var.subscription_id}"
  tenant_id="${var.tenant_id}"
  client_id="${var.client_id}"
  client_secret="${var.client_secret}"
}
`
const ResourceGroupTemplate = `
resource "azurerm_resource_group" "test" {
  name     = "${var.env_id}-test"
  location = "West US"

  tags {
    environment = "Production"
  }
}

`
