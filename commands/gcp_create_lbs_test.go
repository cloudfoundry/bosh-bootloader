package commands_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

const expectedConcourseTemplate = `variable "project_id" {
	type = "string"
}

variable "region" {
	type = "string"
}

variable "zone" {
	type = "string"
}

variable "env_id" {
	type = "string"
}

variable "credentials" {
	type = "string"
}

provider "google" {
	credentials = "${file("${var.credentials}")}"
	project = "${var.project_id}"
	region = "${var.region}"
}

output "external_ip" {
    value = "${google_compute_address.bosh-external-ip.address}"
}

output "network_name" {
    value = "${google_compute_network.bbl-network.name}"
}

output "subnetwork_name" {
    value = "${google_compute_subnetwork.bbl-subnet.name}"
}

output "bosh_open_tag_name" {
    value = "${google_compute_firewall.bosh-open.name}"
}

output "internal_tag_name" {
    value = "${google_compute_firewall.internal.name}"
}

output "director_address" {
	value = "https://${google_compute_address.bosh-external-ip.address}:25555"
}

resource "google_compute_network" "bbl-network" {
  name		 = "${var.env_id}-network"
}

resource "google_compute_subnetwork" "bbl-subnet" {
  name			= "${var.env_id}-subnet"
  ip_cidr_range = "10.0.0.0/16"
  network		= "${google_compute_network.bbl-network.self_link}"
}

resource "google_compute_address" "bosh-external-ip" {
  name = "${var.env_id}-bosh-external-ip"
}

resource "google_compute_firewall" "bosh-open" {
  name    = "${var.env_id}-bosh-open"
  network = "${google_compute_network.bbl-network.name}"

  source_ranges = ["0.0.0.0/0"]

  allow {
    protocol = "icmp"
  }

  allow {
    ports = ["22", "6868", "25555"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-open"]
}

resource "google_compute_firewall" "internal" {
  name    = "${var.env_id}-internal"
  network = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
  }

  allow {
    protocol = "udp"
  }

  source_tags = ["${var.env_id}-bosh-open","${var.env_id}-internal"]
}

output "concourse_target_pool" {
	value = "${google_compute_target_pool.target-pool.name}"
}

resource "google_compute_firewall" "firewall-concourse" {
  name    = "${var.env_id}-concourse-open"
  network = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["443", "2222"]
  }

  target_tags = ["concourse"]
}

resource "google_compute_address" "concourse-address" {
  name = "${var.env_id}-concourse"
}

resource "google_compute_http_health_check" "health-check" {
  name               = "${var.env_id}-concourse"
  request_path       = "/login"
  port               = 443
  check_interval_sec  = 30
  timeout_sec         = 5
  healthy_threshold   = 10
  unhealthy_threshold = 2
}

resource "google_compute_target_pool" "target-pool" {
  name = "${var.env_id}-concourse"

  health_checks = [
    "${google_compute_http_health_check.health-check.name}",
  ]
}

resource "google_compute_forwarding_rule" "ssh-forwarding-rule" {
  name        = "${var.env_id}-concourse-ssh"
  target      = "${google_compute_target_pool.target-pool.self_link}"
  port_range  = "2222"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.concourse-address.address}"
}

resource "google_compute_forwarding_rule" "https-forwarding-rule" {
  name        = "${var.env_id}-concourse-https"
  target      = "${google_compute_target_pool.target-pool.self_link}"
  port_range  = "443"
  ip_protocol = "TCP"
  ip_address  = "${google_compute_address.concourse-address.address}"
}
`

const expectedCFTemplate = `variable "project_id" {
	type = "string"
}

variable "region" {
	type = "string"
}

variable "zone" {
	type = "string"
}

variable "env_id" {
	type = "string"
}

variable "credentials" {
	type = "string"
}

provider "google" {
	credentials = "${file("${var.credentials}")}"
	project = "${var.project_id}"
	region = "${var.region}"
}

output "external_ip" {
    value = "${google_compute_address.bosh-external-ip.address}"
}

output "network_name" {
    value = "${google_compute_network.bbl-network.name}"
}

output "subnetwork_name" {
    value = "${google_compute_subnetwork.bbl-subnet.name}"
}

output "bosh_open_tag_name" {
    value = "${google_compute_firewall.bosh-open.name}"
}

output "internal_tag_name" {
    value = "${google_compute_firewall.internal.name}"
}

output "director_address" {
	value = "https://${google_compute_address.bosh-external-ip.address}:25555"
}

resource "google_compute_network" "bbl-network" {
  name		 = "${var.env_id}-network"
}

resource "google_compute_subnetwork" "bbl-subnet" {
  name			= "${var.env_id}-subnet"
  ip_cidr_range = "10.0.0.0/16"
  network		= "${google_compute_network.bbl-network.self_link}"
}

resource "google_compute_address" "bosh-external-ip" {
  name = "${var.env_id}-bosh-external-ip"
}

resource "google_compute_firewall" "bosh-open" {
  name    = "${var.env_id}-bosh-open"
  network = "${google_compute_network.bbl-network.name}"

  source_ranges = ["0.0.0.0/0"]

  allow {
    protocol = "icmp"
  }

  allow {
    ports = ["22", "6868", "25555"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-open"]
}

resource "google_compute_firewall" "internal" {
  name    = "${var.env_id}-internal"
  network = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
  }

  allow {
    protocol = "udp"
  }

  source_tags = ["${var.env_id}-bosh-open","${var.env_id}-internal"]
}

variable "ssl_certificate" {
  type = "string"
}

variable "ssl_certificate_private_key" {
  type = "string"
}

output "router_backend_service" {
  value = "${google_compute_backend_service.router-lb-backend-service.name}"
}

resource "google_compute_firewall" "firewall-cf" {
  name       = "${var.env_id}-cf-open"
  depends_on = ["google_compute_network.bbl-network"]
  network    = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]

  target_tags = ["${google_compute_backend_service.router-lb-backend-service.name}"]
}

resource "google_compute_global_address" "cf-address" {
  name = "${var.env_id}-cf"
}

resource "google_compute_global_forwarding_rule" "cf-http-forwarding-rule" {
  name       = "${var.env_id}-cf-http"
  ip_address = "${google_compute_global_address.cf-address.address}"
  target     = "${google_compute_target_http_proxy.cf-http-lb-proxy.self_link}"
  port_range = "80"
}

resource "google_compute_global_forwarding_rule" "cf-https-forwarding-rule" {
  name       = "${var.env_id}-cf-https"
  ip_address = "${google_compute_global_address.cf-address.address}"
  target     = "${google_compute_target_https_proxy.cf-https-lb-proxy.self_link}"
  port_range = "443"
}

resource "google_compute_target_http_proxy" "cf-http-lb-proxy" {
  name        = "${var.env_id}-http-proxy"
  description = "really a load balancer but listed as an http proxy"
  url_map     = "${google_compute_url_map.cf-https-lb-url-map.self_link}"
}

resource "google_compute_target_https_proxy" "cf-https-lb-proxy" {
  name             = "${var.env_id}-https-proxy"
  description      = "really a load balancer but listed as an https proxy"
  url_map          = "${google_compute_url_map.cf-https-lb-url-map.self_link}"
  ssl_certificates = ["${google_compute_ssl_certificate.cf-cert.self_link}"]
}

resource "google_compute_ssl_certificate" "cf-cert" {
  name        = "${var.env_id}-lb-cert"
  description = "user provided ssl private key / ssl certificate pair"
  private_key = "${file(var.ssl_certificate_private_key)}"
  certificate = "${file(var.ssl_certificate)}"
}

resource "google_compute_url_map" "cf-https-lb-url-map" {
  name = "${var.env_id}-cf-http"

  default_service = "${google_compute_backend_service.router-lb-backend-service.self_link}"
}

resource "google_compute_http_health_check" "cf-public-health-check" {
  name                = "${var.env_id}-cf"
  port                = 8080
  request_path        = "/health"
  check_interval_sec  = 30
  timeout_sec         = 5
  healthy_threshold   = 10
  unhealthy_threshold = 2
}

resource "google_compute_firewall" "cf-health-check" {
  name       = "${var.env_id}-cf-health-check"
  depends_on = ["google_compute_network.bbl-network"]
  network    = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "tcp"
    ports    = ["8080"]
  }

  source_ranges = ["130.211.0.0/22"]
  target_tags   = ["${google_compute_backend_service.router-lb-backend-service.name}"]
}

resource "google_compute_instance_group" "router-lb-0" {
  name        = "${var.env_id}-router-some-zone"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "some-zone"
}

resource "google_compute_instance_group" "router-lb-1" {
  name        = "${var.env_id}-router-some-other-zone"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "some-other-zone"
}

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

  health_checks = ["${google_compute_http_health_check.cf-public-health-check.self_link}"]
}
`

var _ = Describe("GCPCreateLBs", func() {
	var (
		cloudConfigGenerator *fakes.GCPCloudConfigGenerator
		terraformExecutor    *fakes.TerraformExecutor
		terraformOutputter   *fakes.TerraformOutputter
		boshClientProvider   *fakes.BOSHClientProvider
		boshClient           *fakes.BOSHClient
		zones                *fakes.Zones
		stateStore           *fakes.StateStore
		logger               *fakes.Logger
		command              commands.GCPCreateLBs
		certPath             string
		keyPath              string
		certificate          string
		key                  string
	)

	BeforeEach(func() {
		terraformExecutor = &fakes.TerraformExecutor{}
		cloudConfigGenerator = &fakes.GCPCloudConfigGenerator{}
		terraformOutputter = &fakes.TerraformOutputter{}
		boshClientProvider = &fakes.BOSHClientProvider{}
		boshClient = &fakes.BOSHClient{}
		boshClientProvider.ClientCall.Returns.Client = boshClient
		zones = &fakes.Zones{}
		stateStore = &fakes.StateStore{}
		logger = &fakes.Logger{}

		command = commands.NewGCPCreateLBs(terraformExecutor, terraformOutputter, cloudConfigGenerator, boshClientProvider, zones, stateStore, logger)

		tempCertFile, err := ioutil.TempFile("", "cert")
		Expect(err).NotTo(HaveOccurred())

		certificate = "some-cert"
		certPath = tempCertFile.Name()
		err = ioutil.WriteFile(certPath, []byte(certificate), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		tempKeyFile, err := ioutil.TempFile("", "key")
		Expect(err).NotTo(HaveOccurred())

		key = "some-key"
		keyPath = tempKeyFile.Name()
		err = ioutil.WriteFile(keyPath, []byte(key), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		commands.ResetMarshal()
	})

	Describe("Execute", func() {
		Context("terraform apply call", func() {
			Context("when called with the concourse lb type", func() {
				It("creates and applies a concourse target pool", func() {
					err := command.Execute(commands.GCPCreateLBsConfig{
						LBType: "concourse",
					}, storage.State{
						IAAS:    "gcp",
						EnvID:   "some-env-id",
						TFState: "some-prev-tf-state",
						GCP: storage.GCP{
							ServiceAccountKey: "some-service-account-key",
							Zone:              "some-zone",
							Region:            "some-region",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.StepCall.Messages).To(ContainSequence([]string{
						"generating terraform template", "finished applying terraform template",
					}))
					Expect(terraformExecutor.ApplyCall.CallCount).To(Equal(1))
					Expect(terraformExecutor.ApplyCall.Receives.Credentials).To(Equal("some-service-account-key"))
					Expect(terraformExecutor.ApplyCall.Receives.EnvID).To(Equal("some-env-id"))
					Expect(terraformExecutor.ApplyCall.Receives.Zone).To(Equal("some-zone"))
					Expect(terraformExecutor.ApplyCall.Receives.Region).To(Equal("some-region"))
					Expect(terraformExecutor.ApplyCall.Receives.TFState).To(Equal("some-prev-tf-state"))
					Expect(terraformExecutor.ApplyCall.Receives.Template).To(Equal(expectedConcourseTemplate))
				})
			})

			Context("when called with a cf lb type", func() {
				It("creates and applies a cf backend service", func() {
					zones.GetCall.Returns.Zones = []string{"some-zone", "some-other-zone"}

					err := command.Execute(commands.GCPCreateLBsConfig{
						LBType:   "cf",
						CertPath: certPath,
						KeyPath:  keyPath,
					}, storage.State{
						IAAS:    "gcp",
						EnvID:   "some-env-id",
						TFState: "some-prev-tf-state",
						GCP: storage.GCP{
							ServiceAccountKey: "some-service-account-key",
							Zone:              "some-zone",
							Region:            "some-region",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.StepCall.Messages).To(ContainSequence([]string{
						"generating terraform template", "finished applying terraform template",
					}))
					Expect(terraformExecutor.ApplyCall.CallCount).To(Equal(1))
					Expect(terraformExecutor.ApplyCall.Receives.Credentials).To(Equal("some-service-account-key"))
					Expect(terraformExecutor.ApplyCall.Receives.EnvID).To(Equal("some-env-id"))
					Expect(terraformExecutor.ApplyCall.Receives.Zone).To(Equal("some-zone"))
					Expect(terraformExecutor.ApplyCall.Receives.Region).To(Equal("some-region"))
					Expect(terraformExecutor.ApplyCall.Receives.TFState).To(Equal("some-prev-tf-state"))
					Expect(terraformExecutor.ApplyCall.Receives.Cert).To(Equal(certificate))
					Expect(terraformExecutor.ApplyCall.Receives.Key).To(Equal(key))
					Expect(terraformExecutor.ApplyCall.Receives.Template).To(Equal(expectedCFTemplate))
				})
			})

			It("saves the tf state even if the applier fails", func() {
				expectedError := terraform.NewTerraformApplyError("some-tf-state", errors.New("failed to apply"))
				terraformExecutor.ApplyCall.Returns.Error = expectedError

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{
					IAAS:    "gcp",
					EnvID:   "some-env-id",
					TFState: "some-prev-tf-state",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						Zone:              "some-zone",
						Region:            "some-region",
					},
				})

				Expect(err).To(MatchError("failed to apply"))
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives.State.TFState).To(Equal("some-tf-state"))
			})
		})

		Context("when creating a concourse lb", func() {
			It("creates a cloud-config with concourse lb vm extension", func() {
				terraformOutputter.GetCall.Stub = func(output string) (string, error) {
					switch output {
					case "network_name":
						return "some-network-name", nil
					case "subnetwork_name":
						return "some-subnetwork-name", nil
					case "internal_tag_name":
						return "some-internal-tag", nil
					case "concourse_target_pool":
						return "env-id-concourse-target-pool", nil
					default:
						return "", nil
					}
				}

				zones.GetCall.Returns.Zones = []string{"region1", "region2"}

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						Region: "some-region",
					},
					BOSH: storage.BOSH{
						DirectorUsername: "some-director-username",
						DirectorPassword: "some-director-password",
						DirectorAddress:  "some-director-address",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(zones.GetCall.CallCount).To(Equal(1))
				Expect(zones.GetCall.Receives.Region).To(Equal("some-region"))

				Expect(terraformOutputter.GetCall.CallCount).To(Equal(4))

				Expect(cloudConfigGenerator.GenerateCall.CallCount).To(Equal(1))
				Expect(cloudConfigGenerator.GenerateCall.Receives.CloudConfigInput.AZs).To(Equal([]string{"region1", "region2"}))
				Expect(cloudConfigGenerator.GenerateCall.Receives.CloudConfigInput.Tags).To(Equal([]string{"some-internal-tag"}))
				Expect(cloudConfigGenerator.GenerateCall.Receives.CloudConfigInput.NetworkName).To(Equal("some-network-name"))
				Expect(cloudConfigGenerator.GenerateCall.Receives.CloudConfigInput.SubnetworkName).To(Equal("some-subnetwork-name"))
				Expect(cloudConfigGenerator.GenerateCall.Receives.CloudConfigInput.ConcourseTargetPool).To(Equal("env-id-concourse-target-pool"))

				Expect(logger.StepCall.Messages).To(ContainSequence([]string{
					"generating cloud config", "applying cloud config",
				}))
			})
		})

		Context("when creating a cf lb", func() {
			It("creates a cloud-config with cf lb vm extension", func() {
				terraformOutputter.GetCall.Stub = func(output string) (string, error) {
					switch output {
					case "router_backend_service":
						return "env-id-cf-https-lb", nil
					default:
						return "", nil
					}
				}

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: certPath,
					KeyPath:  keyPath,
				}, storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						Region: "some-region",
					},
					BOSH: storage.BOSH{
						DirectorUsername: "some-director-username",
						DirectorPassword: "some-director-password",
						DirectorAddress:  "some-director-address",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(terraformOutputter.GetCall.CallCount).To(Equal(4))

				Expect(cloudConfigGenerator.GenerateCall.CallCount).To(Equal(1))
				Expect(cloudConfigGenerator.GenerateCall.Receives.CloudConfigInput.CFBackendService).To(Equal("env-id-cf-https-lb"))

				Expect(logger.StepCall.Messages).To(ContainSequence([]string{
					"generating cloud config", "applying cloud config",
				}))
			})
		})

		It("uploads a new cloud-config to the bosh director", func() {
			err := command.Execute(commands.GCPCreateLBsConfig{
				LBType: "concourse",
			}, storage.State{
				IAAS: "gcp",
				BOSH: storage.BOSH{
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
					DirectorAddress:  "some-director-address",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))

			Expect(boshClient.UpdateCloudConfigCall.CallCount).To(Equal(1))
		})

		It("no-ops if SkipIfExists is supplied and the LBType does not change", func() {
			err := command.Execute(commands.GCPCreateLBsConfig{
				LBType:       "concourse",
				SkipIfExists: true,
			}, storage.State{
				IAAS: "gcp",
				LB: storage.LB{
					Type: "concourse",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Messages).To(ContainElement(`lb type "concourse" exists, skipping...`))
			Expect(terraformExecutor.ApplyCall.CallCount).To(Equal(0))
			Expect(terraformOutputter.GetCall.CallCount).To(Equal(0))
			Expect(boshClient.UpdateCloudConfigCall.CallCount).To(Equal(0))
		})

		Context("state manipulation", func() {
			It("saves the concourse lb type", func() {
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{
					IAAS: "gcp",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.SetCall.Receives.State.LB.Type).To(Equal("concourse"))
				Expect(stateStore.SetCall.Receives.State.LB.Cert).To(Equal(""))
				Expect(stateStore.SetCall.Receives.State.LB.Key).To(Equal(""))
			})

			It("saves the cf lb type, cert, and key", func() {
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: certPath,
					KeyPath:  keyPath,
				}, storage.State{
					IAAS: "gcp",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.SetCall.Receives.State.LB.Type).To(Equal("cf"))
				Expect(stateStore.SetCall.Receives.State.LB.Cert).To(Equal("some-cert"))
				Expect(stateStore.SetCall.Receives.State.LB.Key).To(Equal("some-key"))
			})

			It("saves the updated tfstate", func() {
				terraformExecutor.ApplyCall.Returns.TFState = "some-new-tfstate"
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{
					IAAS:    "gcp",
					TFState: "some-old-tfstate",
					BOSH: storage.BOSH{
						DirectorUsername: "some-director-username",
						DirectorPassword: "some-director-password",
						DirectorAddress:  "some-director-address",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.SetCall.Receives.State.TFState).To(Equal("some-new-tfstate"))
			})
		})

		Context("failure cases", func() {
			It("returns an error if the command fails to save the certificate in the state", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{fakes.SetCallReturn{Error: errors.New("failed to save state")}}
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: certPath,
					KeyPath:  keyPath,
				}, storage.State{IAAS: "gcp"})

				Expect(err).To(MatchError("failed to save state"))
			})

			It("returns an error if the command fails to save the key in the state", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, fakes.SetCallReturn{Error: errors.New("failed to save state")}}
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: certPath,
					KeyPath:  keyPath,
				}, storage.State{IAAS: "gcp"})

				Expect(err).To(MatchError("failed to save state"))
			})

			It("returns an error if the command fails to read the certificate", func() {
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: "some/fake/path",
				}, storage.State{IAAS: "gcp"})
				Expect(err).To(MatchError("open some/fake/path: no such file or directory"))
			})

			It("returns an error if the command fails to read the key", func() {
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: certPath,
					KeyPath:  "some/fake/path",
				}, storage.State{IAAS: "gcp"})
				Expect(err).To(MatchError("open some/fake/path: no such file or directory"))
			})

			It("returns an error if applier fails with non terraform apply error", func() {
				terraformExecutor.ApplyCall.Returns.Error = errors.New("failed to apply")
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: certPath,
					KeyPath:  keyPath,
				}, storage.State{
					IAAS: "gcp",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).To(MatchError("failed to apply"))
				Expect(stateStore.SetCall.CallCount).To(Equal(0))
			})

			It("returns an error when the lb type is not concourse or cf", func() {
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "some-fake-lb",
				}, storage.State{IAAS: "gcp"})
				Expect(err).To(MatchError(`"some-fake-lb" is not a valid lb type, valid lb types are: concourse, cf`))
			})

			It("returns an error when the terraform executor fails", func() {
				terraformExecutor.ApplyCall.Returns.Error = errors.New("failed to apply terraform")
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "gcp"})

				Expect(err).To(MatchError("failed to apply terraform"))
			})

			It("returns an error when the state store fails to save the state after applying terraform", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{fakes.SetCallReturn{Error: errors.New("failed to save state")}}
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "gcp"})

				Expect(err).To(MatchError("failed to save state"))
			})

			It("returns an error when both the applier fails and state fails to be set", func() {
				expectedError := terraform.NewTerraformApplyError("some-tf-state", errors.New("failed to apply"))
				terraformExecutor.ApplyCall.Returns.Error = expectedError

				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("state failed to be set")}}
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "gcp"})

				Expect(err).To(MatchError("the following errors occurred:\nfailed to apply,\nstate failed to be set"))
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives.State.TFState).To(Equal("some-tf-state"))
			})

			It("returns an error when the state store fails to save the state after writing lb type", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{fakes.SetCallReturn{Error: nil}, fakes.SetCallReturn{Error: errors.New("failed to save state")}}
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "gcp"})

				Expect(err).To(MatchError("failed to save state"))
			})

			DescribeTable("returns an error when we fail to get an output", func(outputName string) {
				terraformOutputter.GetCall.Stub = func(output string) (string, error) {
					if output == outputName {
						return "", errors.New("failed to get output")
					}
					return "", nil
				}

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "gcp"})
				Expect(err).To(MatchError("failed to get output"))
			},
				Entry("failed to get network_name", "network_name"),
				Entry("failed to get subnetwork_name", "subnetwork_name"),
				Entry("failed to get internal_tag_name", "internal_tag_name"),
				Entry("failed to get concourse_target_pool", "concourse_target_pool"),
			)

			It("returns an error when the cloud config fails to be generated", func() {
				cloudConfigGenerator.GenerateCall.Returns.Error = errors.New("failed to generate cloud config")

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "gcp"})
				Expect(err).To(MatchError("failed to generate cloud config"))
			})

			It("returns an error when the cloud-config fails to marshal", func() {
				commands.SetMarshal(func(interface{}) ([]byte, error) {
					return []byte{}, errors.New("failed to marshal")
				})

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "gcp"})
				Expect(err).To(MatchError("failed to marshal"))
			})

			It("returns an error when the cloud config fails to be updated", func() {
				boshClient.UpdateCloudConfigCall.Returns.Error = errors.New("failed to update cloud config")

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "gcp"})
				Expect(err).To(MatchError("failed to update cloud config"))
			})

			It("returns an error when the iaas type is not gcp", func() {
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("iaas type must be gcp"))
			})

			It("returns an error when the BOSH director does not exist", func() {
				boshClient.InfoCall.Returns.Error = errors.New("error with the director")
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{
					IAAS: "gcp",
				})
				Expect(err).To(MatchError("error with the director"))

				Expect(boshClient.InfoCall.CallCount).To(Equal(1))
			})
		})
	})
})
