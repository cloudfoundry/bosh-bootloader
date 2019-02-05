# key pairs
resource "openstack_compute_keypair_v2" "bosh" {
  region     = "${var.region_name}"
  name       = "${var.env_id}-keypair"
}

# networks
resource "openstack_networking_network_v2" "bosh" {
  region         = "${var.region_name}"
  name           = "${var.env_id}-network"
  admin_state_up = "true"
}

resource "openstack_networking_subnet_v2" "bosh_subnet" {
  region           = "${var.region_name}"
  network_id       = "${openstack_networking_network_v2.bosh.id}"
  cidr             = "10.0.1.0/24"
  ip_version       = 4
  name             = "${var.env_id}-subnet"
  allocation_pools = {
    start = "10.0.1.200"
    end   = "10.0.1.254"
  }
  gateway_ip       = "10.0.1.1"
  enable_dhcp      = "true"
  dns_nameservers  = "${var.dns_nameservers}"
}

# router
resource "openstack_networking_router_v2" "bosh_router" {
  region           = "${var.region_name}"
  name             = "${var.env_id}-router"
  admin_state_up   = "true"
  external_network_id = "${var.ext_net_id}"
}

resource "openstack_networking_router_interface_v2" "bosh_port" {
  region    = "${var.region_name}"
  router_id = "${openstack_networking_router_v2.bosh_router.id}"
  subnet_id = "${openstack_networking_subnet_v2.bosh_subnet.id}"
}

# floating ips
resource "openstack_networking_floatingip_v2" "jb" {
  region = "${var.region_name}"
  pool   = "${var.ext_net_name}"
}

# security groups
resource "openstack_networking_secgroup_v2" "jb" {
  region = "${var.region_name}"
  name = "${var.env_id}-jb"
  description = "Jumpbox Security Group"
}
resource "openstack_networking_secgroup_rule_v2" "jb_rule_1" {
  description = "Anyone to Jumpbox SSH"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 22
  port_range_max = 22
  remote_ip_prefix = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.jb.id}"
}
resource "openstack_networking_secgroup_rule_v2" "jb_rule_2" {
  description = "Anyone to Jumpbox Agent (for bosh create-env -d jumpbox)"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 6868
  port_range_max = 6868
  remote_ip_prefix = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.jb.id}"
}

resource "openstack_networking_secgroup_v2" "bosh" {
  description = "BOSH Director Security Group"
  region = "${var.region_name}"
  name = "${var.env_id}-bosh"
}
resource "openstack_networking_secgroup_rule_v2" "bosh_rule_1" {
  description = "Jumpbox to Director NGINX"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 25555
  port_range_max = 25555
  remote_group_id = "${openstack_networking_secgroup_v2.jb.id}"
  security_group_id = "${openstack_networking_secgroup_v2.bosh.id}"
}
resource "openstack_networking_secgroup_rule_v2" "bosh_rule_2" {
  description = "Jumpbox to Director MBUS"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 6868
  port_range_max = 6868
  remote_group_id = "${openstack_networking_secgroup_v2.jb.id}"
  security_group_id = "${openstack_networking_secgroup_v2.bosh.id}"
}
resource "openstack_networking_secgroup_rule_v2" "bosh_rule_3" {
  description = "Jumpbox to Director SSH"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 22
  port_range_max = 22
  remote_group_id = "${openstack_networking_secgroup_v2.jb.id}"
  security_group_id = "${openstack_networking_secgroup_v2.bosh.id}"
}
resource "openstack_networking_secgroup_rule_v2" "bosh_rule_4" {
  description = "Jumpbox to Director UAA"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 8443
  port_range_max = 8443
  remote_group_id = "${openstack_networking_secgroup_v2.jb.id}"
  security_group_id = "${openstack_networking_secgroup_v2.bosh.id}"
}
resource "openstack_networking_secgroup_rule_v2" "bosh_rule_5" {
  description = "Jumpbox to Director Credhub"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 8844
  port_range_max = 8844
  remote_group_id = "${openstack_networking_secgroup_v2.jb.id}"
  security_group_id = "${openstack_networking_secgroup_v2.bosh.id}"
}
resource "openstack_networking_secgroup_rule_v2" "bosh_rule_6" {
  description = "BOSH deployed VMs to Director NATS"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 4222
  port_range_max = 4222
  remote_group_id = "${openstack_networking_secgroup_v2.vms.id}"
  security_group_id = "${openstack_networking_secgroup_v2.bosh.id}"
}
resource "openstack_networking_secgroup_rule_v2" "bosh_rule_7" {
  description = "BOSH deployed VMs to Director Registry"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 25777
  port_range_max = 25777
  remote_group_id = "${openstack_networking_secgroup_v2.vms.id}"
  security_group_id = "${openstack_networking_secgroup_v2.bosh.id}"
}
resource "openstack_networking_secgroup_rule_v2" "bosh_rule_8" {
  description = "BOSH deployed VMs to Director Blobstore"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 25250
  port_range_max = 25250
  remote_group_id = "${openstack_networking_secgroup_v2.vms.id}"
  security_group_id = "${openstack_networking_secgroup_v2.bosh.id}"
}

resource "openstack_networking_secgroup_v2" "vms" {
  region = "${var.region_name}"
  name = "${var.env_id}-vms"
  description = "BOSH deployed VMs Security Group"
}
resource "openstack_networking_secgroup_rule_v2" "vms_rule_1" {
  description = "BOSH deployed VMs to BOSH deployed VMs TCP"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  remote_group_id = "${openstack_networking_secgroup_v2.vms.id}"
  security_group_id = "${openstack_networking_secgroup_v2.vms.id}"
}
resource "openstack_networking_secgroup_rule_v2" "vms_rule_2" {
  description = "BOSH deployed VMs to BOSH deployed VMs UDP"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "udp"
  remote_group_id = "${openstack_networking_secgroup_v2.vms.id}"
  security_group_id = "${openstack_networking_secgroup_v2.vms.id}"
}
resource "openstack_networking_secgroup_rule_v2" "vms_rule_3" {
  description = "Jumpbox to BOSH deployed VMs SSH"
  region = "${var.region_name}"
  direction = "ingress"
  ethertype = "IPv4"
  protocol = "tcp"
  port_range_min = 22
  port_range_max = 22
  remote_group_id = "${openstack_networking_secgroup_v2.jb.id}"
  security_group_id = "${openstack_networking_secgroup_v2.vms.id}"
}

