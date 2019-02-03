output "default_key_name" {
  value = "${openstack_compute_keypair_v2.bosh.name}"
}

output "private_key" {
  value = "${openstack_compute_keypair_v2.bosh.private_key}"
  sensitive = true
}

output "external_ip" {
  value = "${openstack_networking_floatingip_v2.bosh.address}"
}

output "auth_url" { value = "${var.auth_url}" }
output "az" { value = "${var.availability_zone}" }
output "openstack_project" { value = "${var.tenant_name}" }
output "openstack_domain" { value = "${var.domain_name}" }
output "region" { value = "${var.region_name}" }

output "env_id" { value = "${var.env_id}" }
output "director_name" { value = "${var.env_id}" }
output "director__internal_ip" { value = "10.0.1.6" }
output "jumpbox__internal_ip" { value = "10.0.1.5" }
output "jumpbox_url" { value = "${openstack_networking_floatingip_v2.bosh.address}:22" }

