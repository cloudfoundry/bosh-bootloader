variable "existing-bbl-network" {
  type = string
}

variable "existing-bbl-subnet" {
  type = string
}

variable "existing-bastion-address" {
  type = string
}

resource "google_compute_network" "bbl-network" {
  count = 0
}

resource "google_compute_subnetwork" "bbl-subnet" {
  count = 0
}

resource "google_compute_address" "jumpbox-ip" {
  count = 0
}

data "google_compute_network" "bbl-network" {
  name = "${var.existing-bbl-network}"
}

data "google_compute_subnetwork" "bbl-subnet" {
  name = "${var.existing-bbl-subnet}"
}

data "google_compute_address" "jumpbox-ip" {
  name = "${var.existing-bastion-address}"
}

resource "google_compute_firewall" "external" {
  network = "${data.google_compute_network.bbl-network.name}"
}

resource "google_compute_firewall" "bosh-open" {
  network       = "${data.google_compute_network.bbl-network.name}"
  source_ranges = ["${data.google_compute_address.jumpbox-ip.address}/32"]

  allow {
    ports    = ["22", "6868", "8443", "8844", "25555"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-director"]
}

resource "google_compute_firewall" "bosh-director" {
  network = "${data.google_compute_network.bbl-network.name}"
}

resource "google_compute_firewall" "internal-to-director" {
  network = "${data.google_compute_network.bbl-network.name}"
}

resource "google_compute_firewall" "jumpbox-to-all" {
  network = "${data.google_compute_network.bbl-network.name}"
}

resource "google_compute_firewall" "internal" {
  network = "${data.google_compute_network.bbl-network.name}"
}

output "network" {
  value = "${data.google_compute_network.bbl-network.name}"
}

output "subnetwork" {
  value = "${data.google_compute_subnetwork.bbl-subnet.name}"
}

output "jumpbox_url" {
  value = "${data.google_compute_address.jumpbox-ip.address}"
}

output "director_address" {
  value = "https://${data.google_compute_address.jumpbox-ip.address}:25555"
}

output "external_ip" {
  value = "${data.google_compute_address.jumpbox-ip.address}"
}
