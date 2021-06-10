variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "zone" {
  type = string
}

variable "env_id" {
  type = string
}

variable "credentials" {
  type = string
}

provider "google" {
  credentials = "${file("${var.credentials}")}"
  project     = "${var.project_id}"
  region      = "${var.region}"
}

variable "subnet_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

resource "google_compute_network" "bbl-network" {
  name                    = "${var.env_id}-network"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "bbl-subnet" {
  name          = "${var.env_id}-subnet"
  ip_cidr_range = "${var.subnet_cidr}"
  network       = "${google_compute_network.bbl-network.self_link}"
}

resource "google_compute_firewall" "external" {
  name    = "${var.env_id}-external"
  network = "${google_compute_network.bbl-network.name}"

  source_ranges = ["0.0.0.0/0"]

  allow {
    ports    = ["22", "6868", "25555"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-open"]
}

resource "google_compute_firewall" "bosh-open" {
  name    = "${var.env_id}-bosh-open"
  network = "${google_compute_network.bbl-network.name}"

  source_tags = ["${var.env_id}-bosh-open"]

  allow {
    ports    = ["22", "6868", "8443", "8844", "25555"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-director"]
}

resource "google_compute_firewall" "bosh-director" {
  name    = "${var.env_id}-bosh-director"
  network = "${google_compute_network.bbl-network.name}"

  source_tags = ["${var.env_id}-bosh-director"]

  allow {
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-internal"]
}

resource "google_compute_firewall" "internal-to-director" {
  name    = "${var.env_id}-internal-to-director"
  network = "${google_compute_network.bbl-network.name}"

  source_tags = ["${var.env_id}-internal"]

  allow {
    ports    = ["4222", "25250", "25777"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-director"]
}

resource "google_compute_firewall" "jumpbox-to-all" {
  name    = "${var.env_id}-jumpbox-to-all"
  network = "${google_compute_network.bbl-network.name}"

  source_tags = ["${var.env_id}-jumpbox"]

  allow {
    ports    = ["22", "3389"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-internal", "${var.env_id}-bosh-director"]
}

resource "google_compute_firewall" "internal" {
  name    = "${var.env_id}-internal"
  network = "${google_compute_network.bbl-network.name}"

  source_tags = ["${var.env_id}-internal"]

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
  }

  allow {
    protocol = "udp"
  }

  target_tags = ["${var.env_id}-internal"]
}

locals {
  internal_cidr = "${cidrsubnet(var.subnet_cidr, 8, 0)}"
}

output "network" {
  value = "${google_compute_network.bbl-network.name}"
}

output "subnetwork" {
  value = "${google_compute_subnetwork.bbl-subnet.name}"
}

output "director_name" {
  value = "bosh-${var.env_id}"
}

output "internal_cidr" {
  value = "${local.internal_cidr}"
}

output "internal_gw" {
  value = "${cidrhost(local.internal_cidr, 1)}"
}

output "jumpbox__internal_ip" {
  value = "${cidrhost(local.internal_cidr, 5)}"
}

output "director__internal_ip" {
  value = "${cidrhost(local.internal_cidr, 6)}"
}

output "jumpbox__tags" {
  value = [
    "${google_compute_firewall.bosh-open.name}",
    "${var.env_id}-jumpbox",
  ]
}

output "director__tags" {
  value = ["${google_compute_firewall.bosh-director.name}"]
}

output "internal_tag_name" {
  value = "${google_compute_firewall.internal.name}"
}

resource "google_compute_address" "jumpbox-ip" {
  name = "${var.env_id}-jumpbox-ip"
}

output "jumpbox_url" {
  value = "${google_compute_address.jumpbox-ip.address}:22"
}

output "external_ip" {
  value = "${google_compute_address.jumpbox-ip.address}"
}

output "director_address" {
  value = "https://${google_compute_address.jumpbox-ip.address}:25555"
}

variable "ssl_certificate" {
  type = string
}

variable "ssl_certificate_private_key" {
  type = string
}

output "router_backend_service" {
  value = "${google_compute_backend_service.router-lb-backend-service.name}"
}

output "router_lb_ip" {
  value = "${google_compute_global_address.cf-address.address}"
}

output "ssh_proxy_lb_ip" {
  value = "${google_compute_address.cf-ssh-proxy.address}"
}

output "tcp_router_lb_ip" {
  value = "${google_compute_address.cf-tcp-router.address}"
}

output "ws_lb_ip" {
  value = "${google_compute_address.cf-ws.address}"
}

resource "google_compute_firewall" "firewall-cf" {
  name       = "${var.env_id}-cf-open"
  depends_on = ["google_compute_network.bbl-network"]
  network    = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]

  target_tags = ["${google_compute_backend_service.router-lb-backend-service.name}"]
}

resource "google_compute_global_address" "cf-address" {
  name = "${var.env_id}-cf"
}

resource "google_compute_global_forwarding_rule" "cf-http-forwarding-rule" {
  name       = "${var.env_id}-cf-http"
  ip_address = "${google_compute_global_address.cf-address.address}"
  target     = "${google_compute_target_http_proxy.cf-http-lb-proxy.self_link}"
  port_range = "80"
}

resource "google_compute_global_forwarding_rule" "cf-https-forwarding-rule" {
  name       = "${var.env_id}-cf-https"
  ip_address = "${google_compute_global_address.cf-address.address}"
  target     = "${google_compute_target_https_proxy.cf-https-lb-proxy.self_link}"
  port_range = "443"
}

resource "google_compute_target_http_proxy" "cf-http-lb-proxy" {
  name        = "${var.env_id}-http-proxy"
  description = "really a load balancer but listed as an http proxy"
  url_map     = "${google_compute_url_map.cf-https-lb-url-map.self_link}"
}

resource "google_compute_target_https_proxy" "cf-https-lb-proxy" {
  name             = "${var.env_id}-https-proxy"
  description      = "really a load balancer but listed as an https proxy"
  url_map          = "${google_compute_url_map.cf-https-lb-url-map.self_link}"
  ssl_certificates = ["${google_compute_ssl_certificate.cf-cert.self_link}"]
}

resource "google_compute_ssl_certificate" "cf-cert" {
  name_prefix = "${var.env_id}"
  description = "user provided ssl private key / ssl certificate pair"
  private_key = "${var.ssl_certificate_private_key}"
  certificate = "${var.ssl_certificate}"

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_url_map" "cf-https-lb-url-map" {
  name = "${var.env_id}-cf-http"

  default_service = "${google_compute_backend_service.router-lb-backend-service.self_link}"
}

resource "google_compute_health_check" "cf-public-health-check" {
  name = "${var.env_id}-cf-public"

  http_health_check {
    port         = 8080
    request_path = "/health"
  }
}

resource "google_compute_http_health_check" "cf-public-health-check" {
  name         = "${var.env_id}-cf"
  port         = 8080
  request_path = "/health"
}

resource "google_compute_firewall" "cf-health-check" {
  name       = "${var.env_id}-cf-health-check"
  depends_on = ["google_compute_network.bbl-network"]
  network    = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["8080", "80"]
  }

  source_ranges = ["130.211.0.0/22", "35.191.0.0/16"]
  target_tags   = ["${google_compute_backend_service.router-lb-backend-service.name}"]
}

output "ssh_proxy_target_pool" {
  value = "${google_compute_target_pool.cf-ssh-proxy.name}"
}

resource "google_compute_address" "cf-ssh-proxy" {
  name = "${var.env_id}-cf-ssh-proxy"
}

resource "google_compute_firewall" "cf-ssh-proxy" {
  name       = "${var.env_id}-cf-ssh-proxy-open"
  depends_on = ["google_compute_network.bbl-network"]
  network    = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["2222"]
  }

  target_tags = ["${google_compute_target_pool.cf-ssh-proxy.name}"]
}

resource "google_compute_target_pool" "cf-ssh-proxy" {
  name = "${var.env_id}-cf-ssh-proxy"

  session_affinity = "NONE"
}

resource "google_compute_forwarding_rule" "cf-ssh-proxy" {
  name        = "${var.env_id}-cf-ssh-proxy"
  target      = "${google_compute_target_pool.cf-ssh-proxy.self_link}"
  port_range  = "2222"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.cf-ssh-proxy.address}"
}

output "tcp_router_target_pool" {
  value = "${google_compute_target_pool.cf-tcp-router.name}"
}

resource "google_compute_firewall" "cf-tcp-router" {
  name       = "${var.env_id}-cf-tcp-router"
  depends_on = ["google_compute_network.bbl-network"]
  network    = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["1024-32768"]
  }

  target_tags = ["${google_compute_target_pool.cf-tcp-router.name}"]
}

resource "google_compute_address" "cf-tcp-router" {
  name = "${var.env_id}-cf-tcp-router"
}

resource "google_compute_http_health_check" "cf-tcp-router" {
  name         = "${var.env_id}-cf-tcp-router"
  port         = 80
  request_path = "/health"
}

resource "google_compute_target_pool" "cf-tcp-router" {
  name = "${var.env_id}-cf-tcp-router"

  session_affinity = "NONE"

  health_checks = [
    "${google_compute_http_health_check.cf-tcp-router.name}",
  ]
}

resource "google_compute_forwarding_rule" "cf-tcp-router" {
  name        = "${var.env_id}-cf-tcp-router"
  target      = "${google_compute_target_pool.cf-tcp-router.self_link}"
  port_range  = "1024-32768"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.cf-tcp-router.address}"
}

output "ws_target_pool" {
  value = "${google_compute_target_pool.cf-ws.name}"
}

resource "google_compute_address" "cf-ws" {
  name = "${var.env_id}-cf-ws"
}

resource "google_compute_target_pool" "cf-ws" {
  name = "${var.env_id}-cf-ws"

  session_affinity = "NONE"

  health_checks = ["${google_compute_http_health_check.cf-public-health-check.name}"]
}

resource "google_compute_forwarding_rule" "cf-ws-https" {
  name        = "${var.env_id}-cf-ws-https"
  target      = "${google_compute_target_pool.cf-ws.self_link}"
  port_range  = "443"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.cf-ws.address}"
}

resource "google_compute_forwarding_rule" "cf-ws-http" {
  name        = "${var.env_id}-cf-ws-http"
  target      = "${google_compute_target_pool.cf-ws.self_link}"
  port_range  = "80"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.cf-ws.address}"
}

resource "google_compute_instance_group" "router-lb-0" {
  name        = "${var.env_id}-router-lb-0-us-east1-b"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "us-east1-b"

  named_port {
    name = "https"
    port = "443"
  }
}

resource "google_compute_instance_group" "router-lb-1" {
  name        = "${var.env_id}-router-lb-1-us-east1-c"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "us-east1-c"

  named_port {
    name = "https"
    port = "443"
  }
}

resource "google_compute_instance_group" "router-lb-2" {
  name        = "${var.env_id}-router-lb-2-us-east1-d"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "us-east1-d"

  named_port {
    name = "https"
    port = "443"
  }
}

resource "google_compute_backend_service" "router-lb-backend-service" {
  name        = "${var.env_id}-router-lb"
  port_name   = "https"
  protocol    = "HTTPS"
  timeout_sec = 900
  enable_cdn  = false

  backend {
    group = "${google_compute_instance_group.router-lb-0.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.router-lb-1.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.router-lb-2.self_link}"
  }

  health_checks = ["${google_compute_health_check.cf-public-health-check.self_link}"]
}

output "subnet_cidr_1" {
  value = "${cidrsubnet(var.subnet_cidr, 8, 16)}"
}

output "subnet_cidr_2" {
  value = "${cidrsubnet(var.subnet_cidr, 8, 32)}"
}

output "subnet_cidr_3" {
  value = "${cidrsubnet(var.subnet_cidr, 8, 48)}"
}
