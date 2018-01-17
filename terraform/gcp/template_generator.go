package gcp

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type templates struct {
	vars             string
	jumpbox          string
	boshDirector     string
	cfLB             string
	cfDNS            string
	cfInstanceGroups string
	concourseLB      string
}

type TemplateGenerator struct{}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	tmpls := readTemplates()

	cidrs := t.GenerateSubnetCidrs(state.GCP.Zones)
	template := strings.Join([]string{tmpls.vars, tmpls.boshDirector, tmpls.jumpbox, cidrs}, "\n")

	switch state.LB.Type {
	case "concourse":
		template = strings.Join([]string{template, tmpls.concourseLB}, "\n")
	case "cf":
		backendService := t.GenerateBackendService(state.GCP.Zones)
		template = strings.Join([]string{template, tmpls.cfLB, tmpls.cfInstanceGroups, backendService}, "\n")

		if state.LB.Domain != "" {
			template = strings.Join([]string{template, tmpls.cfDNS}, "\n")
		}
	}

	return template
}

func (t TemplateGenerator) GenerateBackendService(zoneList []string) string {
	backendBaseRestricted := `resource "google_compute_backend_service" "router-lb-backend-service-restricted" {
  count       = "${var.restrict_instance_groups}"
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
`

	backendBase := `resource "google_compute_backend_service" "router-lb-backend-service" {
  count       = "${1 - var.restrict_instance_groups}"
  name        = "${var.env_id}-router-lb"
  port_name   = "https"
  protocol    = "HTTPS"
  timeout_sec = 900
  enable_cdn  = false
%s
  health_checks = ["${google_compute_health_check.cf-public-health-check.self_link}"]
}
`
	var backends string
	for i := 0; i < len(zoneList); i++ {
		backends = fmt.Sprintf(`%s
  backend {
    group = "${google_compute_instance_group.router-lb-%d.self_link}"
  }
`, backends, i)
	}

	return strings.Join([]string{backendBaseRestricted, fmt.Sprintf(backendBase, backends)}, "\n")
}

func (t TemplateGenerator) GenerateSubnetCidrs(zoneList []string) string {
	var cidrs []string
	for i := 0; i < len(zoneList); i++ {
		cidrs = append(cidrs, fmt.Sprintf(`output "subnet_cidr_%d" {
  value = "${cidrsubnet(var.subnet_cidr, 8, %d)}"
}
`, i+1, (i+1)*16))
	}
	return strings.Join(cidrs, "\n")
}

func readTemplates() templates {
	tmpls := templates{}
	tmpls.vars = string(MustAsset("templates/vars.tf"))
	tmpls.jumpbox = string(MustAsset("templates/jumpbox.tf"))
	tmpls.boshDirector = string(MustAsset("templates/bosh_director.tf"))
	tmpls.cfLB = string(MustAsset("templates/cf_lb.tf"))
	tmpls.cfDNS = string(MustAsset("templates/cf_dns.tf"))
	tmpls.cfInstanceGroups = string(MustAsset("templates/cf_instance_groups.tf"))
	tmpls.concourseLB = string(MustAsset("templates/concourse_lb.tf"))

	return tmpls
}
