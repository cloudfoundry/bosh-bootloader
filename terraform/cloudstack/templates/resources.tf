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

resource "cloudstack_network" "bosh_subnet" {
  cidr             = cidrsubnet(var.vpc_cidr, 8, 0)
  name             = "${var.short_env_id}-bosh-subnet"
  vpc_id           = cloudstack_vpc.vpc.id
  display_text     = "bosh-subnet"
  network_offering = var.network_vpc_offering
  zone             = var.cloudstack_zone
  acl_id = var.secure ? element(
    concat(cloudstack_network_acl.bosh_subnet_sec_group.*.id, [""]),
    0,
  ) : cloudstack_network_acl.allow_all.id
}

resource "cloudstack_network" "compilation_subnet" {
  cidr             = cidrsubnet(var.vpc_cidr, 8, 1)
  name             = "${var.short_env_id}-compilation-subnet"
  vpc_id           = cloudstack_vpc.vpc.id
  display_text     = "compilation-subnet"
  network_offering = var.network_vpc_offering
  zone             = var.cloudstack_zone
  acl_id = var.secure ? element(
    concat(cloudstack_network_acl.bosh_subnet_sec_group.*.id, [""]),
    0,
  ) : cloudstack_network_acl.allow_all.id
}

resource "cloudstack_network" "control_plane" {
  cidr             = cidrsubnet(var.vpc_cidr, 6, 1)
  name             = "${var.short_env_id}-control-plane"
  vpc_id           = cloudstack_vpc.vpc.id
  display_text     = "control-plane"
  network_offering = var.network_vpc_offering
  zone             = var.cloudstack_zone
  acl_id = var.secure ? element(
    concat(cloudstack_network_acl.control_plane_sec_group.*.id, [""]),
    0,
  ) : cloudstack_network_acl.allow_all.id
}

resource "cloudstack_network" "data_plane" {
  cidr             = cidrsubnet(var.vpc_cidr, 6, 2)
  name             = "${var.short_env_id}-data-plane"
  vpc_id           = cloudstack_vpc.vpc.id
  display_text     = "data-plane"
  network_offering = var.network_vpc_offering
  zone             = var.cloudstack_zone
  acl_id = var.secure ? element(
    concat(cloudstack_network_acl.data_plane_sec_group.*.id, [""]),
    0,
  ) : cloudstack_network_acl.allow_all.id
}

resource "cloudstack_network" "data_plane_public" {
  count            = var.iso_segment ? 1 : 0
  cidr             = cidrsubnet(var.vpc_cidr, 6, 3)
  name             = "${var.short_env_id}-data-plane-public"
  vpc_id           = cloudstack_vpc.vpc.id
  display_text     = "data-plane-public"
  network_offering = var.network_vpc_offering
  zone             = var.cloudstack_zone
  acl_id = var.secure ? element(
    concat(cloudstack_network_acl.data_plane_sec_group.*.id, [""]),
    0,
  ) : cloudstack_network_acl.allow_all.id
}

resource "cloudstack_ipaddress" "jumpbox_eip" {
  vpc_id = cloudstack_vpc.vpc.id
  zone   = var.cloudstack_zone
}
