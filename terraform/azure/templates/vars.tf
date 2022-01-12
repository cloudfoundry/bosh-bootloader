variable "env_id" {}

variable "region" {}

variable "simple_env_id" {}

variable "subscription_id" {}

variable "tenant_id" {}

variable "client_id" {}

variable "client_secret" {}

variable "network_cidr" {
  default = "10.0.0.0/16"
}

variable "internal_cidr" {
  default = "10.0.0.0/16"
}

provider "azurerm" {
  subscription_id = var.subscription_id
  tenant_id       = var.tenant_id
  client_id       = var.client_id
  client_secret   = var.client_secret

  version = "~> 1.22"
}

provider "tls" {
  version = "~> 1.2"
}

provider "random" {
  version = "~> 2.0"
}
