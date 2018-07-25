package gcp_test

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/gcp"
	"github.com/pmezard/go-difflib/difflib"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TemplateGenerator", func() {
	var (
		templateGenerator gcp.TemplateGenerator
		expectedTemplate  string
		backendService    string
		instanceGroups    string
		zones             []string
		state             storage.State
	)

	BeforeEach(func() {
		templateGenerator = gcp.NewTemplateGenerator()
		zones = []string{"z1", "z2", "z3"}
		instanceGroups = `resource "google_compute_instance_group" "router-lb-0" {
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
`
		backendService = `resource "google_compute_backend_service" "router-lb-backend-service" {
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

  backend {
    group = "${google_compute_instance_group.router-lb-2.self_link}"
  }

  health_checks = ["${google_compute_health_check.cf-public-health-check.self_link}"]
}
`
	})

	Describe("Generate", func() {
		Context("when no lb type is provided", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("vars", "bosh_director", "jumpbox")
			})
			It("uses the base template", func() {
				template := templateGenerator.Generate(storage.State{})
				checkTemplate(template, expectedTemplate)
			})
		})

		Context("when a concourse LB is provided", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("vars", "bosh_director", "jumpbox", "concourse_lb")
				state = storage.State{LB: storage.LB{Type: "concourse"}}
			})
			It("adds the concourse lb template", func() {
				template := templateGenerator.Generate(state)
				checkTemplate(template, expectedTemplate)
			})
		})

		Context("when a CF LB is provided", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("vars", "bosh_director", "jumpbox", "cf_lb")
				expectedTemplate += "\n" + instanceGroups + "\n" + backendService
				state = storage.State{
					GCP: storage.GCP{Zones: []string{"z1", "z2", "z3"}},
					LB:  storage.LB{Type: "cf"},
				}
			})
			It("adds the cf lb template with instance groups and backend service", func() {
				template := templateGenerator.Generate(state)
				checkTemplate(template, expectedTemplate)
			})
		})

		Context("when a CF LB is provided with a domain", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("vars", "bosh_director", "jumpbox", "cf_lb")
				dns := expectTemplate("cf_dns")
				expectedTemplate += "\n" + instanceGroups + "\n" + backendService + "\n" + dns

				state = storage.State{
					GCP: storage.GCP{Zones: []string{"z1", "z2", "z3"}},
					LB: storage.LB{
						Type:   "cf",
						Domain: "some-domain",
					},
				}
			})
			It("adds the cf lb template with instance groups and backend service", func() {
				template := templateGenerator.Generate(state)
				checkTemplate(template, expectedTemplate)
			})
		})
	})

	Describe("GenerateBackendService", func() {
		It("returns a backend service terraform template", func() {
			template := templateGenerator.GenerateBackendService(zones)
			Expect(template).To(Equal(string(backendService)))
		})
	})

	Describe("GenerateInstanceGroups", func() {
		It("returns a backend service terraform template", func() {
			template := templateGenerator.GenerateInstanceGroups(zones)
			Expect(template).To(Equal(string(instanceGroups)))
		})
	})

})

func expectTemplate(parts ...string) string {
	var contents []string
	for _, p := range parts {
		content, err := ioutil.ReadFile(fmt.Sprintf("templates/%s.tf", p))
		Expect(err).NotTo(HaveOccurred())
		contents = append(contents, string(content))
	}
	return strings.Join(contents, "\n")
}

func checkTemplate(actual, expected string) {
	if actual != string(expected) {
		diff, _ := difflib.GetContextDiffString(difflib.ContextDiff{
			A:        difflib.SplitLines(actual),
			B:        difflib.SplitLines(string(expected)),
			FromFile: "actual",
			ToFile:   "expected",
			Context:  10,
		})
		fmt.Println(diff)
	}

	Expect(actual).To(Equal(string(expected)))
}
