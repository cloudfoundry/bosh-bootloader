variable "env_id" {}

variable "region" {
  type = "string"
}

variable "simple_env_id" {}

variable "subscription_id" {}

variable "tenant_id" {}

variable "client_id" {}

variable "client_secret" {}

variable "network_cidr" {
  default = "10.0.0.0/16"
}

variable "internal_cidr" {
  default = "10.0.0.0/24"
}

provider "azurerm" {
  subscription_id = "${var.subscription_id}"
  tenant_id       = "${var.tenant_id}"
  client_id       = "${var.client_id}"
  client_secret   = "${var.client_secret}"
}
