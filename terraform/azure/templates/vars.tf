variable "env_id" {
	type = "string"
}

variable "location" {
	type = "string"
}

variable "simple_env_id" {
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

variable "internal_cidr" {
  type    = "string"
  default = "10.0.0.0/24"
}

provider "azurerm" {
  subscription_id  = "${var.subscription_id}"
  tenant_id        = "${var.tenant_id}"
  client_id        = "${var.client_id}"
  client_secret    = "${var.client_secret}"
}
