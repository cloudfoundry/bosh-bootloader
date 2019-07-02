variable "cf_env_count" {
  default = 1
}

output "router_backend_service" {
  value = "${google_compute_backend_service.router-lb-backend-service.0.name}"
}

output "router_backend_services" {
  value = "${google_compute_backend_service.router-lb-backend-service.*.name}"
}

output "router_lb_ip" {
  value = "${google_compute_global_address.cf-address.0.address}"
}

output "ssh_proxy_lb_ip" {
  value = "${google_compute_address.cf-ssh-proxy.0.address}"
}

output "tcp_router_lb_ip" {
  value = "${google_compute_address.cf-tcp-router.0.address}"
}

output "ws_lb_ip" {
  value = "${google_compute_address.cf-ws.0.address}"
}

resource "google_compute_instance_group" "router-lb-a" {
  name        = "${var.env_id}-router-lb-us-west1-a-${count.index}"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "us-west1-a"

  named_port {
    name = "https"
    port = "443"
  }

  count = "${var.cf_env_count}"
}

resource "google_compute_instance_group" "router-lb-b" {
  name        = "${var.env_id}-router-lb-1-us-west1-b-${count.index}"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "us-west1-b"

  named_port {
    name = "https"
    port = "443"
  }

  count = "${var.cf_env_count}"
}

resource "google_compute_backend_service" "router-lb-backend-service" {
  name        = "${var.env_id}-router-lb-${count.index}"
  port_name   = "https"
  protocol    = "HTTPS"
  timeout_sec = 900
  enable_cdn  = false

  backend {
    group = "${element(google_compute_instance_group.router-lb-a.*.self_link, count.index)}"
  }

  backend {
    group = "${element(google_compute_instance_group.router-lb-b.*.self_link, count.index)}"
  }

  health_checks = ["${element(google_compute_health_check.cf-public-health-check.*.self_link, count.index)}"]

  count = "${var.cf_env_count}"
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

  target_tags = ["${google_compute_backend_service.router-lb-backend-service.*.name}"]
}

resource "google_compute_global_address" "cf-address" {
  name = "${var.env_id}-cf-${count.index}"

  count = "${var.cf_env_count}"
}

resource "google_compute_global_forwarding_rule" "cf-http-forwarding-rule" {
  name       = "${var.env_id}-cf-http-${count.index}"
  ip_address = "${element(google_compute_global_address.cf-address.*.address, count.index)}"
  target     = "${element(google_compute_target_http_proxy.cf-http-lb-proxy.*.self_link, count.index)}"
  port_range = "80"

  count = "${var.cf_env_count}"
}

resource "google_compute_global_forwarding_rule" "cf-https-forwarding-rule" {
  name       = "${var.env_id}-cf-https-${count.index}"
  ip_address = "${element(google_compute_global_address.cf-address.*.address, count.index)}"
  target     = "${element(google_compute_target_https_proxy.cf-https-lb-proxy.*.self_link, count.index)}"
  port_range = "443"

  count = "${var.cf_env_count}"
}

resource "google_compute_target_http_proxy" "cf-http-lb-proxy" {
  name        = "${var.env_id}-http-proxy-${count.index}"
  description = "really a load balancer but listed as an http proxy"
  url_map     = "${element(google_compute_url_map.cf-https-lb-url-map.*.self_link, count.index)}"

  count = "${var.cf_env_count}"
}

resource "google_compute_target_https_proxy" "cf-https-lb-proxy" {
  name             = "${var.env_id}-https-proxy-${count.index}"
  description      = "really a load balancer but listed as an https proxy"
  url_map          = "${element(google_compute_url_map.cf-https-lb-url-map.*.self_link, count.index)}"
  ssl_certificates = ["${google_compute_ssl_certificate.cf-cert.self_link}"]

  count = "${var.cf_env_count}"
}

resource "google_compute_url_map" "cf-https-lb-url-map" {
  name = "${var.env_id}-cf-http-${count.index}"

  default_service = "${element(google_compute_backend_service.router-lb-backend-service.*.self_link, count.index)}"

  count = "${var.cf_env_count}"
}

resource "google_compute_health_check" "cf-public-health-check" {
  name = "${var.env_id}-cf-public-${count.index}"

  http_health_check {
    port         = 8080
    request_path = "/health"
  }

  count = "${var.cf_env_count}"
}

resource "google_compute_http_health_check" "cf-public-health-check" {
  name         = "${var.env_id}-cf-${count.index}"
  port         = 8080
  request_path = "/health"

  count = "${var.cf_env_count}"
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
  target_tags   = ["${google_compute_backend_service.router-lb-backend-service.*.name}"]
}

output "ssh_proxy_target_pool" {
  value = "${google_compute_target_pool.cf-ssh-proxy.0.name}"
}

output "ssh_proxy_target_pools" {
  value = "${google_compute_target_pool.cf-ssh-proxy.*.name}"
}

resource "google_compute_address" "cf-ssh-proxy" {
  name = "${var.env_id}-cf-ssh-proxy-${count.index}"

  count = "${var.cf_env_count}"
}

resource "google_compute_firewall" "cf-ssh-proxy" {
  name       = "${var.env_id}-cf-ssh-proxy-open"
  depends_on = ["google_compute_network.bbl-network"]
  network    = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["2222"]
  }

  target_tags = ["${google_compute_target_pool.cf-ssh-proxy.*.name}"]
}

resource "google_compute_target_pool" "cf-ssh-proxy" {
  name = "${var.env_id}-cf-ssh-proxy-${count.index}"

  session_affinity = "NONE"

  count = "${var.cf_env_count}"
}

resource "google_compute_forwarding_rule" "cf-ssh-proxy" {
  name        = "${var.env_id}-cf-ssh-proxy-${count.index}"
  target      = "${element(google_compute_target_pool.cf-ssh-proxy.*.self_link, count.index)}"
  port_range  = "2222"
  ip_protocol = "TCP"
  ip_address  = "${element(google_compute_address.cf-ssh-proxy.*.address, count.index)}"

  count = "${var.cf_env_count}"
}

output "tcp_router_target_pool" {
  value = "${google_compute_target_pool.cf-tcp-router.0.name}"
}

output "tcp_router_target_pools" {
  value = "${google_compute_target_pool.cf-tcp-router.*.name}"
}

resource "google_compute_firewall" "cf-tcp-router" {
  name       = "${var.env_id}-cf-tcp-router"
  depends_on = ["google_compute_network.bbl-network"]
  network    = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["1024-32768"]
  }

  target_tags = ["${google_compute_target_pool.cf-tcp-router.*.name}"]
}

resource "google_compute_address" "cf-tcp-router" {
  name = "${var.env_id}-cf-tcp-router-${count.index}"

  count = "${var.cf_env_count}"
}

resource "google_compute_http_health_check" "cf-tcp-router" {
  name         = "${var.env_id}-cf-tcp-router-${count.index}"
  port         = 80
  request_path = "/health"

  count = "${var.cf_env_count}"
}

resource "google_compute_target_pool" "cf-tcp-router" {
  name = "${var.env_id}-cf-tcp-router-${count.index}"

  session_affinity = "NONE"

  health_checks = [
    "${element(google_compute_http_health_check.cf-tcp-router.*.name, count.index)}",
  ]

  count = "${var.cf_env_count}"
}

resource "google_compute_forwarding_rule" "cf-tcp-router" {
  name        = "${var.env_id}-cf-tcp-router-${count.index}"
  target      = "${element(google_compute_target_pool.cf-tcp-router.*.self_link, count.index)}"
  port_range  = "1024-32768"
  ip_protocol = "TCP"
  ip_address  = "${element(google_compute_address.cf-tcp-router.*.address, count.index)}"

  count = "${var.cf_env_count}"
}

output "ws_target_pool" {
  value = "${google_compute_target_pool.cf-ws.0.name}"
}

output "ws_target_pools" {
  value = "${google_compute_target_pool.cf-ws.*.name}"
}

resource "google_compute_address" "cf-ws" {
  name = "${var.env_id}-cf-ws-${count.index}"

  count = "${var.cf_env_count}"
}

resource "google_compute_target_pool" "cf-ws" {
  name = "${var.env_id}-cf-ws-${count.index}"
  session_affinity = "NONE"
  health_checks = ["${element(google_compute_http_health_check.cf-public-health-check.*.name, count.index)}"]

  count = "${var.cf_env_count}"
}

resource "google_compute_forwarding_rule" "cf-ws-https" {
  name        = "${var.env_id}-cf-ws-https-${count.index}"
  target      = "${element(google_compute_target_pool.cf-ws.*.self_link, count.index)}"
  port_range  = "443"
  ip_protocol = "TCP"
  ip_address  = "${element(google_compute_address.cf-ws.*.address, count.index)}"

  count = "${var.cf_env_count}"
}

resource "google_compute_forwarding_rule" "cf-ws-http" {
  name        = "${var.env_id}-cf-ws-http-${count.index}"
  target      = "${element(google_compute_target_pool.cf-ws.*.self_link, count.index)}"
  port_range  = "80"
  ip_protocol = "TCP"
  ip_address  = "${element(google_compute_address.cf-ws.*.address, count.index)}"

  count = "${var.cf_env_count}"
}
