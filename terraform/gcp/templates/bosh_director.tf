variable "subnet_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

resource "google_compute_network" "bbl-network" {
  name                    = "${var.env_id}-network"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "bbl-subnet" {
  name          = "${var.env_id}-subnet"
  ip_cidr_range = "${var.subnet_cidr}"
  network       = "${google_compute_network.bbl-network.self_link}"
}

resource "google_compute_firewall" "external" {
  name    = "${var.env_id}-external"
  network = "${google_compute_network.bbl-network.name}"

  source_ranges = ["0.0.0.0/0"]

  allow {
    ports    = ["22", "6868", "25555"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-open"]
}

resource "google_compute_firewall" "bosh-open" {
  name    = "${var.env_id}-bosh-open"
  network = "${google_compute_network.bbl-network.name}"

  source_tags = ["${var.env_id}-bosh-open"]

  allow {
    ports    = ["22", "6868", "8443", "8844", "25555"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-director"]
}

resource "google_compute_firewall" "bosh-director" {
  name    = "${var.env_id}-bosh-director"
  network = "${google_compute_network.bbl-network.name}"

  source_tags = ["${var.env_id}-bosh-director"]

  allow {
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-internal"]
}

resource "google_compute_firewall" "internal-to-director" {
  name    = "${var.env_id}-internal-to-director"
  network = "${google_compute_network.bbl-network.name}"

  source_tags = ["${var.env_id}-internal"]

  allow {
    ports    = ["4222", "25250", "25777"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-director"]
}

resource "google_compute_firewall" "jumpbox-to-all" {
  name    = "${var.env_id}-jumpbox-to-all"
  network = "${google_compute_network.bbl-network.name}"

  source_tags = ["${var.env_id}-jumpbox"]

  allow {
    ports    = ["22", "3389"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-internal", "${var.env_id}-bosh-director"]
}

resource "google_compute_firewall" "internal" {
  name    = "${var.env_id}-internal"
  network = "${google_compute_network.bbl-network.name}"

  source_tags = ["${var.env_id}-internal"]

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
  }

  allow {
    protocol = "udp"
  }

  target_tags = ["${var.env_id}-internal"]
}

output "network" {
  value = "${google_compute_network.bbl-network.name}"
}

output "subnetwork" {
  value = "${google_compute_subnetwork.bbl-subnet.name}"
}

output "internal_cidr" {
  value = "${var.subnet_cidr}"
}

output "internal_gw" {
  value = "${google_compute_subnetwork.bbl-subnet.gateway_address}"
}

output "director_name" {
  value = "bosh-${var.env_id}"
}

output "jumpbox__internal_ip" {
  value = "${cidrhost(var.subnet_cidr, 5)}"
}

output "director__internal_ip" {
  value = "${cidrhost(var.subnet_cidr, 6)}"
}

output "jumpbox__tags" {
  value = [
    "${google_compute_firewall.bosh-open.name}",
    "${var.env_id}-jumpbox",
  ]
}

output "director__tags" {
  value = ["${google_compute_firewall.bosh-director.name}"]
}

output "internal_tag_name" {
  value = "${google_compute_firewall.internal.name}"
}
