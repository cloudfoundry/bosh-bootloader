# key pairs
resource "openstack_compute_keypair_v2" "bosh" {
  region     = "${var.region_name}"
  name       = "bosh-${var.tenant_name}"
  # public_key = "${replace("${file("../bosh.pub")}","\n","")}"
}

# security groups
resource "openstack_networking_secgroup_v2" "secgroup" {
  region = "${var.region_name}"
  name = "bosh${var.security_group_suffix}"
  description = "BOSH Security Group"
}

resource "openstack_networking_secgroup_rule_v2" "secgroup_rule_4" {
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 22
  port_range_max = 22
  remote_ip_prefix = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.secgroup.id}"
}

resource "openstack_networking_secgroup_rule_v2" "secgroup_rule_6" {
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 6868
  port_range_max = 6868
  remote_ip_prefix = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.secgroup.id}"
}

resource "openstack_networking_secgroup_rule_v2" "secgroup_rule_5" {
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 25555
  port_range_max = 25555
  remote_ip_prefix = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.secgroup.id}"
}

resource "openstack_networking_secgroup_rule_v2" "secgroup_rule_1" {
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  remote_group_id = "${openstack_networking_secgroup_v2.secgroup.id}"
  security_group_id = "${openstack_networking_secgroup_v2.secgroup.id}"
}

# floating ips
resource "openstack_networking_floatingip_v2" "bosh" {
  region = "${var.region_name}"
  pool   = "${var.ext_net_name}"
}
