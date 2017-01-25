package commands

const terraformVarsTemplate = `variable "project_id" {
	type = "string"
}

variable "region" {
	type = "string"
}

variable "zone" {
	type = "string"
}

variable "env_id" {
	type = "string"
}

variable "credentials" {
	type = "string"
}

provider "google" {
	credentials = "${file("${var.credentials}")}"
	project = "${var.project_id}"
	region = "${var.region}"
}
`

const terraformBOSHDirectorTemplate = `output "external_ip" {
    value = "${google_compute_address.bosh-external-ip.address}"
}

output "network_name" {
    value = "${google_compute_network.bbl-network.name}"
}

output "subnetwork_name" {
    value = "${google_compute_subnetwork.bbl-subnet.name}"
}

output "bosh_open_tag_name" {
    value = "${google_compute_firewall.bosh-open.name}"
}

output "internal_tag_name" {
    value = "${google_compute_firewall.internal.name}"
}

output "director_address" {
	value = "https://${google_compute_address.bosh-external-ip.address}:25555"
}

resource "google_compute_network" "bbl-network" {
  name		 = "${var.env_id}-network"
}

resource "google_compute_subnetwork" "bbl-subnet" {
  name			= "${var.env_id}-subnet"
  ip_cidr_range = "10.0.0.0/16"
  network		= "${google_compute_network.bbl-network.self_link}"
}

resource "google_compute_address" "bosh-external-ip" {
  name = "${var.env_id}-bosh-external-ip"
}

resource "google_compute_firewall" "bosh-open" {
  name    = "${var.env_id}-bosh-open"
  network = "${google_compute_network.bbl-network.name}"

  source_ranges = ["0.0.0.0/0"]

  allow {
    protocol = "icmp"
  }

  allow {
    ports = ["22", "6868", "25555"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-open"]
}

resource "google_compute_firewall" "internal" {
  name    = "${var.env_id}-internal"
  network = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
  }

  allow {
    protocol = "udp"
  }

  source_tags = ["${var.env_id}-bosh-open","${var.env_id}-internal"]
}
`

const terraformConcourseLBTemplate = `variable "ssl_certificate" {
  type = "string"
}

variable "ssl_certificate_private_key" {
  type = "string"
}

output "web_backend_service" {
  value = "${google_compute_backend_service.concourse-lb-backend-service.name}"
}

output "ssh_target_pool" {
	value = "${google_compute_target_pool.ssh-target-pool.name}"
}

output "concourse_lb_ip" {
    value = "${google_compute_global_address.concourse-address.address}"
}

resource "google_compute_firewall" "firewall-web" {
  name    = "${var.env_id}-concourse-web-open"
  network = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["443"]
  }

  target_tags = ["${google_compute_backend_service.concourse-lb-backend-service.name}"]
}

resource "google_compute_firewall" "firewall-ssh" {
  name    = "${var.env_id}-concourse-ssh-open"
  network = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["2222"]
  }

  target_tags = ["${google_compute_target_pool.ssh-target-pool.name}"]
}

resource "google_compute_global_address" "concourse-address" {
  name = "${var.env_id}-concourse"
}

resource "google_compute_address" "concourse-ssh-address" {
  name = "${var.env_id}-concourse-ssh"
}

resource "google_compute_target_pool" "ssh-target-pool" {
  name = "${var.env_id}-concourse"

  health_checks = ["${google_compute_http_health_check.concourse-health-check.name}"]
}

resource "google_compute_ssl_certificate" "concourse-cert" {
  name_prefix = "${var.env_id}"
  private_key = "${file(var.ssl_certificate_private_key)}"
  certificate = "${file(var.ssl_certificate)}"
  lifecycle {
	create_before_destroy = true
  }
}

resource "google_compute_global_forwarding_rule" "concourse-https-forwarding-rule" {
  name       = "${var.env_id}-concourse-https"
  ip_address = "${google_compute_global_address.concourse-address.address}"
  target     = "${google_compute_target_https_proxy.concourse-https-lb-proxy.self_link}"
  port_range = "443"
}

resource "google_compute_forwarding_rule" "ssh-forwarding-rule" {
  name        = "${var.env_id}-concourse-ssh"
  target      = "${google_compute_target_pool.ssh-target-pool.self_link}"
  port_range  = "2222"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.concourse-ssh-address.address}"
}

resource "google_compute_target_https_proxy" "concourse-https-lb-proxy" {
  name             = "${var.env_id}-https-proxy"
  url_map          = "${google_compute_url_map.concourse-https-lb-url-map.self_link}"
  ssl_certificates = ["${google_compute_ssl_certificate.concourse-cert.self_link}"]
}

resource "google_compute_url_map" "concourse-https-lb-url-map" {
  name = "${var.env_id}-concourse-https"

  default_service = "${google_compute_backend_service.concourse-lb-backend-service.self_link}"
}

resource "google_compute_http_health_check" "concourse-health-check" {
  name               = "${var.env_id}-concourse"
  request_path       = "/login"
  port               = 80
  check_interval_sec  = 30
  timeout_sec         = 5
  healthy_threshold   = 10
  unhealthy_threshold = 2
}

resource "google_compute_firewall" "concourse-health-check" {
  name       = "${var.env_id}-concourse-health-check"
  depends_on = ["google_compute_network.bbl-network"]
  network    = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["80"]
  }

  source_ranges = ["130.211.0.0/22"]
  target_tags   = [
    "${google_compute_backend_service.concourse-lb-backend-service.name}",
    "${google_compute_target_pool.ssh-target-pool.name}"
  ]
}
`

const terraformCFLBTemplate = `variable "ssl_certificate" {
  type = "string"
}

variable "ssl_certificate_private_key" {
  type = "string"
}

output "router_backend_service" {
  value = "${google_compute_backend_service.cf-router-lb-backend-service.name}"
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

  target_tags = ["${google_compute_backend_service.cf-router-lb-backend-service.name}"]
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
  private_key = "${file(var.ssl_certificate_private_key)}"
  certificate = "${file(var.ssl_certificate)}"
  lifecycle {
	create_before_destroy = true
  }
}

resource "google_compute_url_map" "cf-https-lb-url-map" {
  name = "${var.env_id}-cf-http"

  default_service = "${google_compute_backend_service.cf-router-lb-backend-service.self_link}"
}

resource "google_compute_http_health_check" "cf-router-health-check" {
  name                = "${var.env_id}-cf"
  port                = 8080
  request_path        = "/health"
  check_interval_sec  = 30
  timeout_sec         = 5
  healthy_threshold   = 10
  unhealthy_threshold = 2
}

resource "google_compute_firewall" "cf-health-check" {
  name       = "${var.env_id}-cf-health-check"
  depends_on = ["google_compute_network.bbl-network"]
  network    = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["8080"]
  }

  source_ranges = ["130.211.0.0/22"]
  target_tags   = ["${google_compute_backend_service.cf-router-lb-backend-service.name}"]
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
  name                = "${var.env_id}-cf-tcp-router"
  port                = 80
  request_path        = "/health"
  check_interval_sec  = 30
  timeout_sec         = 5
  healthy_threshold   = 10
  unhealthy_threshold = 2
}

resource "google_compute_target_pool" "cf-tcp-router" {
  name = "${var.env_id}-cf-tcp-router"

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

  health_checks = ["${google_compute_http_health_check.cf-router-health-check.name}"]
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
`

const terraformCFDNSTemplate = `
variable "system_domain" {
  type = "string"
}

resource "google_dns_managed_zone" "env_dns_zone" {
  name        = "${var.env_id}-zone"
  dns_name    = "${var.system_domain}."
  description = "DNS zone for the ${var.env_id} environment"
}

resource "google_dns_record_set" "wildcard-dns" {
  name       = "*.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_global_address.cf-address"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_global_address.cf-address.address}"]
}

resource "google_dns_record_set" "bosh-dns" {
  name       = "bosh.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.bosh-external-ip"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.bosh-external-ip.address}"]
}

resource "google_dns_record_set" "cf-ssh-proxy" {
  name       = "ssh.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-ssh-proxy"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.cf-ssh-proxy.address}"]
}

resource "google_dns_record_set" "tcp-dns" {
  name       = "tcp.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-tcp-router"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.cf-tcp-router.address}"]
}

resource "google_dns_record_set" "doppler-dns" {
  name       = "doppler.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-ws"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.cf-ws.address}"]
}

resource "google_dns_record_set" "loggregator-dns" {
  name       = "loggregator.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-ws"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.cf-ws.address}"]
}

resource "google_dns_record_set" "wildcard-ws-dns" {
  name       = "*.ws.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-ws"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.cf-ws.address}"]
}
`
