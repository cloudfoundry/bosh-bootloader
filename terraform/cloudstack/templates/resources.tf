locals {
  director_name        = "bosh-${var.env_id}"
  internal_cidr        = cidrsubnet(var.vpc_cidr, 8, 0)
  internal_gw          = cidrhost(local.internal_cidr, 1)
  jumpbox_internal_ip  = cidrhost(local.internal_cidr, 5)
  director_internal_ip = cidrhost(local.internal_cidr, 6)
}


resource "tls_private_key" "bosh_vms" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "cloudstack_ssh_keypair" "bosh_vms" {
  name       = "${var.env_id}_bosh_vms"
  public_key = tls_private_key.bosh_vms.public_key_openssh
}

resource "cloudstack_vpc" "vpc" {
  name         = "${var.env_id}-vpc"
  cidr         = var.vpc_cidr
  vpc_offering = var.vpc_offering
  zone         = var.cloudstack_zone
}

resource "cloudstack_network_acl" "allow_all" {
  name   = "allow_all"
  vpc_id = cloudstack_vpc.vpc.id
}

resource "cloudstack_network_acl_rule" "ingress_allow_all" {
  acl_id = cloudstack_network_acl.allow_all.id

  rule {
    action = "allow"
    cidr_list = [
      "0.0.0.0/0",
    ]
    protocol     = "all"
    ports        = []
    traffic_type = "ingress"
  }
}

resource "cloudstack_network_acl_rule" "egress_allow_all" {
  acl_id = cloudstack_network_acl.allow_all.id

  rule {
    action = "allow"
    cidr_list = [
      "0.0.0.0/0",
    ]
    protocol     = "all"
    ports        = []
    traffic_type = "egress"
  }
}

resource "cloudstack_ipaddress" "jumpbox_eip" {
  vpc_id = cloudstack_vpc.vpc.id
  zone   = var.cloudstack_zone
}
