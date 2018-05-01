locals {
  short_env_id = "${substr(var.env_id, 0, min(20, length(var.env_id)))}"
}

resource "google_compute_firewall" "bosh-director-lite" {
  name    = "${local.short_env_id}-bosh-director-lite"
  network = "${google_compute_network.bbl-network.name}"

  source_ranges = ["0.0.0.0/0"]

  allow {
    ports    = ["80", "443", "2222"]
    protocol = "tcp"
  }

  target_tags = ["${local.short_env_id}-bosh-director"]
}

resource "google_compute_address" "bosh-director-ip" {
  name = "${var.env_id}-bosh-director-ip"
}

resource "google_compute_route" "bosh-lite-vms" {
  name        = "${var.env_id}-bosh-lite-vms"
  dest_range  = "10.244.0.0/16"
  network     = "${google_compute_network.bbl-network.name}"
  next_hop_ip = "10.0.0.6"
  priority    = 1

  depends_on = ["google_compute_subnetwork.bbl-subnet"]
}

resource "google_compute_firewall" "bosh-director-lite-tcp-routing" {
  name    = "${local.short_env_id}-bosh-director-lite-tcp-routing"
  network = "${google_compute_network.bbl-network.name}"

  source_ranges = ["0.0.0.0/0"]

  allow {
    ports    = ["1024-1123"]
    protocol = "tcp"
  }

  target_tags = ["${local.short_env_id}-bosh-director"]
}

output "external_ip" {
  value = "${google_compute_address.bosh-director-ip.address}"
}

output "jumpbox__external_ip" {
  value = "${google_compute_address.jumpbox-ip.address}"
}
