provider "openstack" {
  auth_url    = "${var.auth_url}"
  user_name   = "${var.user_name}"
  password    = "${var.password}"
  tenant_name = "${var.tenant_name}"
  domain_name = "${var.domain_name}"
  insecure    = "${var.insecure}"
  cacert_file = "${var.cacert_file}"
}

terraform {
  required_providers {
    openstack = {
      source = "terraform-provider-openstack/openstack"
      version = ">= 1.50.0"
    }
  }
}