variable "project_id" {}

variable "region" {}

variable "zone" {}

variable "env_id" {}

variable "credentials" {}

provider "google" {
  credentials = "${file("${var.credentials}")}"
  project = "${var.project_id}"
  region = "${var.region}"
}
