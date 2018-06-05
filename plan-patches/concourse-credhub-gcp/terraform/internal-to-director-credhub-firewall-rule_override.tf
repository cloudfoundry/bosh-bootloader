resource "google_compute_firewall" "internal-to-director-credhub" {
  name    = "${var.env_id}-internal-to-director-credhub"
  network = "${google_compute_network.bbl-network.name}"

  source_tags = ["${var.env_id}-internal"]

  allow {
    ports    = ["8844", "8443"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-director"]
}
