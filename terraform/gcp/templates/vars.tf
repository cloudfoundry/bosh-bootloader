variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "zone" {
  type = string
}

variable "env_id" {
  type = string
}

variable "credentials" {
  type = string
}

provider "google" {
  credentials = "${file("${var.credentials}")}"
  project     = "${var.project_id}"
  region      = "${var.region}"
}
