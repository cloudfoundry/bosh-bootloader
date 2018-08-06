output "concourse_target_pool" {
  value = "${google_compute_target_pool.target-pool.name}"
}

output "concourse_lb_ip" {
  value = "${google_compute_address.concourse-address.address}"
}

resource "google_compute_firewall" "firewall-concourse" {
  name    = "${var.env_id}-concourse-open"
  network = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["80", "443", "2222", "8443", "8844"]
  }

  target_tags = ["concourse"]
}

resource "google_compute_address" "concourse-address" {
  name = "${var.env_id}-concourse"
}

resource "google_compute_target_pool" "target-pool" {
  name = "${var.env_id}-concourse"

  session_affinity = "NONE"
}

resource "google_compute_forwarding_rule" "ssh-forwarding-rule" {
  name        = "${var.env_id}-concourse-ssh"
  target      = "${google_compute_target_pool.target-pool.self_link}"
  port_range  = "2222"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.concourse-address.address}"
}

resource "google_compute_forwarding_rule" "https-forwarding-rule" {
  name        = "${var.env_id}-concourse-https"
  target      = "${google_compute_target_pool.target-pool.self_link}"
  port_range  = "443"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.concourse-address.address}"
}

resource "google_compute_forwarding_rule" "http-forwarding-rule" {
  name        = "${var.env_id}-concourse-http"
  target      = "${google_compute_target_pool.target-pool.self_link}"
  port_range  = "80"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.concourse-address.address}"
}

resource "google_compute_forwarding_rule" "credhub-forwarding-rule" {
  name        = "${var.env_id}-concourse-credhub"
  target      = "${google_compute_target_pool.target-pool.self_link}"
  port_range  = "8844"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.concourse-address.address}"
}

resource "google_compute_forwarding_rule" "uaa-forwarding-rule" {
  name        = "${var.env_id}-concourse-uaa"
  target      = "${google_compute_target_pool.target-pool.self_link}"
  port_range  = "8443"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.concourse-address.address}"
}
