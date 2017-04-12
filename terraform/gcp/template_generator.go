package gcp

import (
	"fmt"
	"strings"
)

type TemplateGenerator struct {
	zones zones
}

type zones interface {
	Get(region string) []string
}

const backendBase = `resource "google_compute_backend_service" "router-lb-backend-service" {
  name        = "${var.env_id}-router-lb"
  port_name   = "http"
  protocol    = "HTTP"
  timeout_sec = 900
  enable_cdn  = false
%s
  health_checks = ["${google_compute_http_health_check.cf-public-health-check.self_link}"]
}
`

func NewTemplateGenerator(zones zones) TemplateGenerator {
	return TemplateGenerator{
		zones: zones,
	}
}

func (t TemplateGenerator) GenerateBackendService(region string) string {
	zones := t.zones.Get(region)
	var backends string
	for i := 0; i < len(zones); i++ {
		backends = fmt.Sprintf(`%s
  backend {
    group = "${google_compute_instance_group.router-lb-%d.self_link}"
  }
`, backends, i)
	}

	return fmt.Sprintf(backendBase, backends)
}

func (t TemplateGenerator) GenerateInstanceGroups(region string) string {
	zones := t.zones.Get(region)
	var groups []string
	for i, zone := range zones {
		groups = append(groups, fmt.Sprintf(`resource "google_compute_instance_group" "router-lb-%[1]d" {
  name        = "${var.env_id}-router-lb-%[1]d-%[2]s"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "%[2]s"
}
`, i, zone))
	}

	return strings.Join(groups, "\n")
}
