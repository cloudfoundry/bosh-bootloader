resource "google_compute_backend_service" "router-lb-backend-service" {
  name        = "${var.env_id}-router-lb"
  port_name   = "https"
  protocol    = "HTTPS"
  timeout_sec = 900
  enable_cdn  = false

  backend {
    group = "${google_compute_instance_group.router-lb-0.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.router-lb-1.self_link}"
  }

  health_checks = ["${google_compute_health_check.cf-public-health-check.self_link}"]
}

resource "google_compute_instance_group" "router-lb-2" {
  count       = "0"
  name        = "${var.env_id}-router-lb-2-z3"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "z3"

  named_port {
    name = "https"
    port = "443"
  }
}

resource "google_compute_instance_group" "router-lb-3" {
  count       = "0"
  name        = "${var.env_id}-router-lb-3-z4"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "z4"

  named_port {
    name = "https"
    port = "443"
  }
}
