variable "project_id" {
  type = "string"
}

variable "region" {
  type = "string"
}

variable "zone" {
  type = "string"
}

variable "env_id" {
  type = "string"
}

variable "credentials" {
  type = "string"
}

variable "zones" {
  type = "list"
}

variable "restrict_instance_groups" {
  default = false
}

variable "subnet_cidr" {
  type    = "string"
  default = "10.0.0.0/16"
}

provider "google" {
  credentials = "${file("${var.credentials}")}"
  project     = "${var.project_id}"
  region      = "${var.region}"
}
