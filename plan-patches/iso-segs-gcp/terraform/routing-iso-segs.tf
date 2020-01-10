output "iso_router_backend_service" {
  value = "${google_compute_backend_service.iso-router-lb-backend-service.name}"
}

resource "google_compute_global_address" "iso-cf-address" {
  name = "${var.env_id}-iso-cf"
}

resource "google_compute_global_forwarding_rule" "iso-cf-http-forwarding-rule" {
  name       = "${var.env_id}-iso-cf-http"
  ip_address = "${google_compute_global_address.iso-cf-address.address}"
  target     = "${google_compute_target_http_proxy.iso-cf-http-lb-proxy.self_link}"
  port_range = "80"
}

resource "google_compute_global_forwarding_rule" "iso-cf-https-forwarding-rule" {
  name       = "${var.env_id}-iso-cf-https"
  ip_address = "${google_compute_global_address.iso-cf-address.address}"
  target     = "${google_compute_target_https_proxy.iso-cf-https-lb-proxy.self_link}"
  port_range = "443"
}

resource "google_compute_target_http_proxy" "iso-cf-http-lb-proxy" {
  name        = "${var.env_id}-iso-http-proxy"
  description = "really a load balancer but listed as an http proxy"
  url_map     = "${google_compute_url_map.iso-cf-https-lb-url-map.self_link}"
}

resource "google_compute_target_https_proxy" "iso-cf-https-lb-proxy" {
  name             = "${var.env_id}-iso-https-proxy"
  description      = "really a load balancer but listed as an https proxy"
  url_map          = "${google_compute_url_map.iso-cf-https-lb-url-map.self_link}"
  ssl_certificates = ["${google_compute_ssl_certificate.cf-cert.self_link}"]
}

resource "google_compute_url_map" "iso-cf-https-lb-url-map" {
  name = "${var.env_id}-iso-cf-http"

  default_service = "${google_compute_backend_service.iso-router-lb-backend-service.self_link}"
}

resource "google_compute_backend_service" "iso-router-lb-backend-service" {
  name        = "${var.env_id}-iso-router-lb"
  port_name   = "https"
  protocol    = "HTTPS"
  timeout_sec = 900
  enable_cdn  = false

  backend {
    group = "${google_compute_instance_group.iso-router-lb-0.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.iso-router-lb-1.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.iso-router-lb-2.self_link}"
  }

  health_checks = ["${google_compute_health_check.cf-public-health-check.self_link}"]
}

resource "google_compute_instance_group" "iso-router-lb-0" {
  name        = "${var.env_id}-iso-router-lb-0-us-central1-a"
  description = "isolation-segment: terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "us-central1-a"

  named_port {
    name = "https"
    port = "443"
  }
}

resource "google_compute_instance_group" "iso-router-lb-1" {
  name        = "${var.env_id}-iso-router-lb-1-us-central1-b"
  description = "isolation-segment: terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "us-central1-b"

  named_port {
    name = "https"
    port = "443"
  }
}

resource "google_compute_instance_group" "iso-router-lb-2" {
  name        = "${var.env_id}-iso-router-lb-2-us-central1-c"
  description = "isolation-segment: terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "us-central1-c"

  named_port {
    name = "https"
    port = "443"
  }
}

output "iso_ws_target_pool" {
  value = "${google_compute_target_pool.iso-cf-ws.name}"
}

resource "google_compute_address" "iso-cf-ws" {
  name = "${var.env_id}-iso-cf-ws"
}

resource "google_compute_target_pool" "iso-cf-ws" {
  name = "${var.env_id}-iso-cf-ws"

  session_affinity = "NONE"

  health_checks = ["${google_compute_http_health_check.cf-public-health-check.name}"]
}

resource "google_compute_forwarding_rule" "iso-cf-ws-https" {
  name        = "${var.env_id}-iso-cf-ws-https"
  target      = "${google_compute_target_pool.iso-cf-ws.self_link}"
  port_range  = "443"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.iso-cf-ws.address}"
}

resource "google_compute_forwarding_rule" "cf-iso-ws-http" {
  name        = "${var.env_id}-iso-cf-ws-http"
  target      = "${google_compute_target_pool.iso-cf-ws.self_link}"
  port_range  = "80"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.iso-cf-ws.address}"
}

resource "google_dns_record_set" "iso-wildcard-dns" {
  name       = "*.iso-seg.${google_dns_managed_zone.env_dns_zone.dns_name}"
  depends_on = ["google_compute_global_address.cf-address"]
  type       = "A"
  ttl        = 300

  managed_zone = "${google_dns_managed_zone.env_dns_zone.name}"

  rrdatas = ["${google_compute_global_address.iso-cf-address.address}"]
}

resource "google_compute_firewall" "iso-firewall-cf" {
  name       = "${var.env_id}-iso-cf-open"
  depends_on = ["google_compute_network.bbl-network"]
  network    = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]

  target_tags = ["${google_compute_backend_service.iso-router-lb-backend-service.name}"]
}

resource "google_compute_firewall" "iso-cf-health-check" {
  name       = "${var.env_id}-iso-cf-health-check"
  depends_on = ["google_compute_network.bbl-network"]
  network    = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["8080", "80"]
  }

  source_ranges = ["130.211.0.0/22", "35.191.0.0/16"]
  target_tags   = ["${google_compute_backend_service.iso-router-lb-backend-service.name}"]
}
