resource "google_compute_firewall" "bosh-director-lite" {
  name = "${var.env_id}-bosh-director-lite"
  network = "${google_compute_network.bbl-network.name}"

  source_ranges = ["0.0.0.0/0"]

  allow {
    ports = ["80", "443", "2222"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-director"]
}

resource "google_compute_address" "bosh-director-ip" {
  name = "${var.env_id}-bosh-director-ip"
}

output "bosh_director_external_ip" {
    value = "${google_compute_address.bosh-director-ip.address}"
}
