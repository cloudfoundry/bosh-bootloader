output "default_key_name" {
  value = "${openstack_compute_keypair_v2.bosh.name}"
}

output "private_key" {
  value = "${openstack_compute_keypair_v2.bosh.private_key}"
  sensitive = true
}

output "external_ip" {
  value = "${openstack_networking_floatingip_v2.jb.address}"
}

output "vms_security_groups" {
  value = ["${openstack_networking_secgroup_v2.vms.name}"]
}

output "jumpbox__default_security_groups" {
  value = ["${openstack_networking_secgroup_v2.jb.name}"]
}

output "director__default_security_groups" {
  value = ["${openstack_networking_secgroup_v2.bosh.name}"]
}

output "internal_cidr" {
  value = "${openstack_networking_subnet_v2.bosh_subnet.cidr}"
}

output "internal_gw" {
  value = "${openstack_networking_subnet_v2.bosh_subnet.gateway_ip}"
}

output "net_id" {
  value = "${openstack_networking_network_v2.bosh.id}"
}

output "internal_ip" {
  value = "${cidrhost(openstack_networking_subnet_v2.bosh_subnet.cidr, 10)}"
}

output "router_id" {
  value = "${openstack_networking_router_v2.bosh_router.id}"
}

output "director__internal_ip" {
  value = "${cidrhost(openstack_networking_subnet_v2.bosh_subnet.cidr, 6)}"
}

output "jumpbox__internal_ip" {
  value = "${cidrhost(openstack_networking_subnet_v2.bosh_subnet.cidr, 5)}"
}

output "jumpbox_url" {
  value = "${openstack_networking_floatingip_v2.jb.address}:22"
}

output "auth_url" {
  value = "${var.auth_url}"
}

output "az" {
  value = "${var.availability_zone}"
}

output "openstack_project" {
  value = "${var.tenant_name}"
}

output "openstack_domain" {
  value = "${var.domain_name}"
}

output "region" {
  value = "${var.region_name}"
}

output "env_id" {
  value = "${var.env_id}"
}

output "director_name" {
  value = "${var.env_id}"
}
