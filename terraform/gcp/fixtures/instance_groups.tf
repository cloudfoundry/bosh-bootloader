resource "google_compute_instance_group" "router-lb-0" {
  name        = "${var.env_id}-router-z1"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "z1"
}

resource "google_compute_instance_group" "router-lb-1" {
  name        = "${var.env_id}-router-z2"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "z2"
}

resource "google_compute_instance_group" "router-lb-2" {
  name        = "${var.env_id}-router-z3"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "z3"
}
