resource "google_compute_address" "cfcr_tcp" {
  name = "${var.env_id}-cfcr"
}

resource "google_compute_target_pool" "cfcr_tcp_public" {
    region = "${var.region}"
    name = "${var.env_id}-cfcr-tcp-public"
}

resource "google_compute_forwarding_rule" "cfcr_tcp" {
  name        = "${var.env_id}-cfcr-tcp"
  target      = "${google_compute_target_pool.cfcr_tcp_public.self_link}"
  port_range  = "8443"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.cfcr_tcp.address}"
}

resource "google_compute_firewall" "cfcr_tcp_public" {
  name    = "${var.env_id}-cfcr-tcp-public"
  network       = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["8443"]
  }

  target_tags = ["master"]
}
