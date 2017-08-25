resource "google_compute_instance_group" "router-lb-0" {
  name        = "${var.env_id}-router-lb-0-z1"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "z1"

  named_port {
    name = "https"
    port = "443"
  }
}

resource "google_compute_instance_group" "router-lb-1" {
  name        = "${var.env_id}-router-lb-1-z2"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "z2"

  named_port {
    name = "https"
    port = "443"
  }
}

resource "google_compute_instance_group" "router-lb-2" {
  name        = "${var.env_id}-router-lb-2-z3"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "z3"

  named_port {
    name = "https"
    port = "443"
  }
}
