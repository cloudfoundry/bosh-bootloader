# networks
resource "openstack_networking_network_v2" "bosh" {
  region         = "${var.region_name}"
  name           = "bosh"
  admin_state_up = "true"
}

resource "openstack_networking_subnet_v2" "bosh_subnet" {
  region           = "${var.region_name}"
  network_id       = "${openstack_networking_network_v2.bosh.id}"
  cidr             = "10.0.1.0/24"
  ip_version       = 4
  name             = "bosh_sub"
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
  name             = "bosh-router"
  admin_state_up   = "true"
  external_network_id = "${var.ext_net_id}"
}

resource "openstack_networking_router_interface_v2" "bosh_port" {
  region    = "${var.region_name}"
  router_id = "${openstack_networking_router_v2.bosh_router.id}"
  subnet_id = "${openstack_networking_subnet_v2.bosh_subnet.id}"
}
