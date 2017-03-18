resource "google_compute_backend_service" "router-lb-backend-service" {
  name        = "${var.env_id}-router-lb"
  port_name   = "http"
  protocol    = "HTTP"
  timeout_sec = 900
  enable_cdn  = false

  backend {
    group = "${google_compute_instance_group.router-lb-0.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.router-lb-1.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.router-lb-2.self_link}"
  }

  health_checks = ["${google_compute_http_health_check.cf-public-health-check.self_link}"]
}
