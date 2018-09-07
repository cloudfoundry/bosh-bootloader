resource "google_compute_global_address" "cf-address" {
  count = 0
}

resource "google_compute_address" "cf-address" {
  name = "${var.env_id}-cf"
}

resource "google_compute_global_forwarding_rule" "cf-http-forwarding-rule" {
  count = 0
}

resource "google_compute_forwarding_rule" "cf-http-forwarding-rule" {
  name        = "${var.env_id}-cf-http"
  ip_address  = "${google_compute_address.cf-address.address}"
  target      = "${google_compute_target_pool.router-lb-target-pool.self_link}"
  port_range  = "80"
  ip_protocol = "TCP"
}

resource "google_compute_global_forwarding_rule" "cf-https-forwarding-rule" {
  count = 0
}

resource "google_compute_forwarding_rule" "cf-https-forwarding-rule" {
  name        = "${var.env_id}-cf-https"
  ip_address  = "${google_compute_address.cf-address.address}"
  target      = "${google_compute_target_pool.router-lb-target-pool.self_link}"
  port_range  = "443"
  ip_protocol = "TCP"
}

resource "google_compute_target_http_proxy" "cf-http-lb-proxy" {
  count = 0
}

resource "google_compute_target_https_proxy" "cf-https-lb-proxy" {
  count = 0
}

resource "google_compute_ssl_certificate" "cf-cert" {
  count = 0
}

resource "google_compute_url_map" "cf-https-lb-url-map" {
  count = 0
}

resource "google_compute_health_check" "cf-public-health-check" {
  count = 0
}

resource "google_compute_backend_service" "router-lb-backend-service" {
  count = 0
}

resource "google_compute_target_pool" "router-lb-target-pool" {
  name = "${var.env_id}-router-lb"

  health_checks = [
    "${google_compute_http_health_check.cf-public-health-check.name}",
  ]
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
  target_tags   = ["${google_compute_target_pool.router-lb-target-pool.name}"]
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

  target_tags = ["${google_compute_target_pool.router-lb-target-pool.name}"]
}

resource "google_dns_record_set" "wildcard-dns" {
  name       = "*.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_address.cf-address"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_address.cf-address.address}"]
}

# Should always be empty, added to avoid a terraform warning
output "router_backend_service" {
  value = ""
}

output "router_target_pool" {
  value = "${google_compute_target_pool.router-lb-target-pool.name}"
}

output "router_lb_ip" {
  value = "${google_compute_address.cf-address.address}"
}

output "ws_lb_ip" {
  value = ""
}

output "ws_target_pool" {
  value = ""
}

resource "google_compute_address" "cf-ws" {
  count = 0
}

resource "google_compute_target_pool" "cf-ws" {
  count = 0
}

resource "google_compute_forwarding_rule" "cf-ws-https" {
  count = 0
}

resource "google_compute_forwarding_rule" "cf-ws-http" {
  count = 0
}

resource "google_dns_record_set" "doppler-dns" {
  count = 0
}

resource "google_dns_record_set" "loggregator-dns" {
  count = 0
}

resource "google_dns_record_set" "wildcard-ws-dns" {
  count = 0
}
