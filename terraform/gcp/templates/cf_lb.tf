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
