output "prometheus_target_pool" {
  value = "${google_compute_target_pool.prometheus-target-pool.name}"
}

output "prometheus_lb_ip" {
  value = "${google_compute_address.prometheus-address.address}"
}

resource "google_compute_firewall" "firewall-prometheus" {
  name    = "${var.env_id}-prometheus-open"
  network = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["3000", "9090", "9093"]
  }

  target_tags = ["prometheus"]
}

resource "google_compute_address" "prometheus-address" {
  name = "${var.env_id}-prometheus"
}

resource "google_compute_target_pool" "prometheus-target-pool" {
  name = "${var.env_id}-prometheus"

  session_affinity = "NONE"
}

resource "google_compute_forwarding_rule" "grafana-forwarding-rule" {
  name        = "${var.env_id}-prometheus-grafana"
  target      = "${google_compute_target_pool.prometheus-target-pool.self_link}"
  port_range  = "3000"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.prometheus-address.address}"
}

resource "google_compute_forwarding_rule" "prometheus-forwarding-rule" {
  name        = "${var.env_id}-prometheus-prometheus"
  target      = "${google_compute_target_pool.prometheus-target-pool.self_link}"
  port_range  = "9090"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.prometheus-address.address}"
}

resource "google_compute_forwarding_rule" "alertmanager-forwarding-rule" {
  name        = "${var.env_id}-prometheus-alertmanager"
  target      = "${google_compute_target_pool.prometheus-target-pool.self_link}"
  port_range  = "9093"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.prometheus-address.address}"
}
