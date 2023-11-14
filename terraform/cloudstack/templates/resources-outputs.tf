output "cloudstack_default_key_name" {
  value = "${var.env_id}_bosh_vms"
}

output "private_key" {
  value     = tls_private_key.bosh_vms.private_key_pem
  sensitive = true
}

output "external_ip" {
  value = cloudstack_ipaddress.jumpbox_eip.ip_address
}

output "jumpbox_url" {
  value = "${cloudstack_ipaddress.jumpbox_eip.ip_address}:22"
}

output "director_address" {
  value = "https://${cloudstack_ipaddress.jumpbox_eip.ip_address}:25555"
}

output "network_name" {
  value = "${var.short_env_id}-bosh-subnet"
}

output "director_name" {
  value = local.director_name
}

output "internal_cidr" {
  value = local.internal_cidr
}

output "internal_gw" {
  value = local.internal_gw
}

output "jumpbox__internal_ip" {
  value = local.jumpbox_internal_ip
}

output "director__internal_ip" {
  value = local.director_internal_ip
}

output "cloudstack_compute_offering" {
  value = var.cloudstack_compute_offering
}

output "cloudstack_zone" {
  value = var.cloudstack_zone
}

output "cloudstack_endpoint" {
  value = var.cloudstack_endpoint
}

output "dns" {
  value = var.dns
}
