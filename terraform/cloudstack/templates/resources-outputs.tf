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

output "internal_subnet_cidr_mapping" {
  value = {
    (cloudstack_network.data_plane.name)    = cloudstack_network.data_plane.cidr,
    (cloudstack_network.control_plane.name) = cloudstack_network.control_plane.cidr,
    (cloudstack_network.bosh_subnet.name)   = cloudstack_network.bosh_subnet.cidr,
    (element(concat(cloudstack_network.data_plane_public.*.name, [
      ""]), 0)) = element(concat(cloudstack_network.data_plane_public.*.cidr, [
    ""]), 0),
    (cloudstack_network.compilation_subnet.name) = cloudstack_network.compilation_subnet.cidr,
  }
}

output "internal_subnet_gw_mapping" {
  value = {
    (cloudstack_network.data_plane.name)    = cloudstack_network.data_plane.gateway,
    (cloudstack_network.control_plane.name) = cloudstack_network.control_plane.gateway,
    (cloudstack_network.bosh_subnet.name)   = cloudstack_network.bosh_subnet.gateway,
    (element(concat(cloudstack_network.data_plane_public.*.name, [""]), 0)) = element(
      concat(cloudstack_network.data_plane_public.*.gateway, [""]),
      0,
    ),
    (cloudstack_network.compilation_subnet.name) = cloudstack_network.compilation_subnet.gateway,
  }
}
