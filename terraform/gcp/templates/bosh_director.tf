variable "subnet_cidr" {
  type    = "string"
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

locals {
  network = "${google_compute_network.bbl-network.name}"
  subnetwork = "${google_compute_subnetwork.bbl-subnet.name}"
  director_name = "bosh-${var.env_id}"
  internal_cidr = "${cidrsubnet(var.subnet_cidr, 8, 0)}"
  internal_gw = "${cidrhost(local.internal_cidr, 1)}"
  jumpbox_internal_ip = "${cidrhost(local.internal_cidr, 5)}"
  director_internal_ip = "${cidrhost(local.internal_cidr, 6)}"
  bosh_open_tag_name = "${google_compute_firewall.bosh-open.name}"
  bosh_director_tag_name = "${google_compute_firewall.bosh-director.name}"
  jumpbox_tag_name = "${var.env_id}-jumpbox"
  internal_tag_name = "${google_compute_firewall.internal.name}"
}

output "network" {
  value = "${local.network}"
}

output "subnetwork" {
  value = "${local.subnetwork}"
}

output "director_name" {
  value = "${local.director_name}"
}

output "internal_cidr" {
  value = "${local.internal_cidr}"
}

output "internal_gw" {
  value = "${local.internal_gw}"
}

output "jumpbox__internal_ip" {
  value = "${local.jumpbox_internal_ip}"
}

output "director__internal_ip" {
  value = "${local.director_internal_ip}"
}

output "jumpbox__tags" {
  value = [
    "${local.bosh_open_tag_name}",
    "${local.jumpbox_tag_name}"
  ]
}

output "director__tags" {
  value = [
    "${local.bosh_director_tag_name}"
  ]
}

output "internal_tag_name" {
  value = "${local.internal_tag_name}"
}
