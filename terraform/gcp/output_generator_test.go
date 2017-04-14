package gcp_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("OutputGenerator", func() {
	var (
		executor        *fakes.TerraformExecutor
		outputGenerator gcp.OutputGenerator
	)

	BeforeEach(func() {
		executor = &fakes.TerraformExecutor{}
		outputGenerator = gcp.NewOutputGenerator(executor)
	})

	Context("when no lb exists", func() {
		BeforeEach(func() {
			executor.OutputCall.Stub = func(output string) (string, error) {
				switch output {
				case "external_ip":
					return "some-external-ip", nil
				case "network_name":
					return "some-network-name", nil
				case "subnetwork_name":
					return "some-subnetwork-name", nil
				case "bosh_open_tag_name":
					return "some-bosh-open-tag-name", nil
				case "internal_tag_name":
					return "some-internal-tag-name", nil
				case "director_address":
					return "some-director-address", nil
				default:
					return "", fmt.Errorf("unexpected output requested: %s", output)
				}
			}
		})

		It("returns all terraform outputs except lb related outputs", func() {
			outputs, err := outputGenerator.Generate(storage.State{
				IAAS:  "gcp",
				EnvID: "some-env-id",
				GCP: storage.GCP{
					ServiceAccountKey: "some-service-account-key",
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "some-region",
				},
				TFState: "some-tf-state",
				LB: storage.LB{
					Type:   "",
					Domain: "",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(outputs).To(Equal(map[string]interface{}{
				"external_ip":        "some-external-ip",
				"network_name":       "some-network-name",
				"subnetwork_name":    "some-subnetwork-name",
				"bosh_open_tag_name": "some-bosh-open-tag-name",
				"internal_tag_name":  "some-internal-tag-name",
				"director_address":   "some-director-address",
			}))
		})
	})

	Context("when cf lb exists", func() {
		Context("when the domain is not specified", func() {
			BeforeEach(func() {
				executor.OutputCall.Stub = func(output string) (string, error) {
					switch output {
					case "external_ip":
						return "some-external-ip", nil
					case "network_name":
						return "some-network-name", nil
					case "subnetwork_name":
						return "some-subnetwork-name", nil
					case "bosh_open_tag_name":
						return "some-bosh-open-tag-name", nil
					case "internal_tag_name":
						return "some-internal-tag-name", nil
					case "director_address":
						return "some-director-address", nil
					case "router_backend_service":
						return "some-router-backend-service", nil
					case "ssh_proxy_target_pool":
						return "some-ssh-proxy-target-pool", nil
					case "tcp_router_target_pool":
						return "some-tcp-router-target-pool", nil
					case "ws_target_pool":
						return "some-ws-target-pool", nil
					case "router_lb_ip":
						return "some-router-lb-ip", nil
					case "ssh_proxy_lb_ip":
						return "some-ssh-proxy-lb-ip", nil
					case "tcp_router_lb_ip":
						return "some-tcp-router-lb-ip", nil
					case "ws_lb_ip":
						return "some-ws-lb-ip", nil
					default:
						return "", fmt.Errorf("unexpected output requested: %s", output)
					}
				}
			})

			It("returns all terraform outputs related to cf lb without system domain DNS servers", func() {
				outputs, err := outputGenerator.Generate(storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
					TFState: "some-tf-state",
					LB: storage.LB{
						Type:   "cf",
						Domain: "",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(outputs).To(Equal(map[string]interface{}{
					"external_ip":            "some-external-ip",
					"network_name":           "some-network-name",
					"subnetwork_name":        "some-subnetwork-name",
					"bosh_open_tag_name":     "some-bosh-open-tag-name",
					"internal_tag_name":      "some-internal-tag-name",
					"director_address":       "some-director-address",
					"router_backend_service": "some-router-backend-service",
					"ssh_proxy_target_pool":  "some-ssh-proxy-target-pool",
					"tcp_router_target_pool": "some-tcp-router-target-pool",
					"ws_target_pool":         "some-ws-target-pool",
					"router_lb_ip":           "some-router-lb-ip",
					"ssh_proxy_lb_ip":        "some-ssh-proxy-lb-ip",
					"tcp_router_lb_ip":       "some-tcp-router-lb-ip",
					"ws_lb_ip":               "some-ws-lb-ip",
				}))
			})
		})

		Context("when the domain is specified", func() {
			BeforeEach(func() {
				executor.OutputCall.Stub = func(output string) (string, error) {
					switch output {
					case "external_ip":
						return "some-external-ip", nil
					case "network_name":
						return "some-network-name", nil
					case "subnetwork_name":
						return "some-subnetwork-name", nil
					case "bosh_open_tag_name":
						return "some-bosh-open-tag-name", nil
					case "internal_tag_name":
						return "some-internal-tag-name", nil
					case "director_address":
						return "some-director-address", nil
					case "router_backend_service":
						return "some-router-backend-service", nil
					case "ssh_proxy_target_pool":
						return "some-ssh-proxy-target-pool", nil
					case "tcp_router_target_pool":
						return "some-tcp-router-target-pool", nil
					case "ws_target_pool":
						return "some-ws-target-pool", nil
					case "router_lb_ip":
						return "some-router-lb-ip", nil
					case "ssh_proxy_lb_ip":
						return "some-ssh-proxy-lb-ip", nil
					case "tcp_router_lb_ip":
						return "some-tcp-router-lb-ip", nil
					case "ws_lb_ip":
						return "some-ws-lb-ip", nil
					case "system_domain_dns_servers":
						return "some-name-server-1,\nsome-name-server-2", nil
					default:
						return "", fmt.Errorf("unexpected output requested: %s", output)
					}
				}
			})

			It("returns terraform outputs related to cf lb with the system domain DNS servers", func() {
				outputs, err := outputGenerator.Generate(storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
					TFState: "some-tf-state",
					LB: storage.LB{
						Type:   "cf",
						Domain: "some-domain",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(outputs).To(Equal(map[string]interface{}{
					"external_ip":               "some-external-ip",
					"network_name":              "some-network-name",
					"subnetwork_name":           "some-subnetwork-name",
					"bosh_open_tag_name":        "some-bosh-open-tag-name",
					"internal_tag_name":         "some-internal-tag-name",
					"director_address":          "some-director-address",
					"router_backend_service":    "some-router-backend-service",
					"ssh_proxy_target_pool":     "some-ssh-proxy-target-pool",
					"tcp_router_target_pool":    "some-tcp-router-target-pool",
					"ws_target_pool":            "some-ws-target-pool",
					"router_lb_ip":              "some-router-lb-ip",
					"ssh_proxy_lb_ip":           "some-ssh-proxy-lb-ip",
					"tcp_router_lb_ip":          "some-tcp-router-lb-ip",
					"ws_lb_ip":                  "some-ws-lb-ip",
					"system_domain_dns_servers": []string{"some-name-server-1", "some-name-server-2"},
				}))
			})
		})
	})

	Context("when concourse lb exists", func() {
		BeforeEach(func() {
			executor.OutputCall.Stub = func(output string) (string, error) {
				switch output {
				case "external_ip":
					return "some-external-ip", nil
				case "network_name":
					return "some-network-name", nil
				case "subnetwork_name":
					return "some-subnetwork-name", nil
				case "bosh_open_tag_name":
					return "some-bosh-open-tag-name", nil
				case "internal_tag_name":
					return "some-internal-tag-name", nil
				case "director_address":
					return "some-director-address", nil
				case "concourse_target_pool":
					return "some-concourse-target-pool", nil
				case "concourse_lb_ip":
					return "some-concourse-lb-ip", nil
				default:
					return "", fmt.Errorf("unexpected output requested: %s", output)
				}
			}
		})

		It("returns terraform outputs related to concourse lb", func() {
			outputs, err := outputGenerator.Generate(storage.State{
				IAAS:  "gcp",
				EnvID: "some-env-id",
				GCP: storage.GCP{
					ServiceAccountKey: "some-service-account-key",
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "some-region",
				},
				TFState: "some-tf-state",
				LB: storage.LB{
					Type:   "concourse",
					Domain: "",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(outputs).To(Equal(map[string]interface{}{
				"external_ip":           "some-external-ip",
				"network_name":          "some-network-name",
				"subnetwork_name":       "some-subnetwork-name",
				"bosh_open_tag_name":    "some-bosh-open-tag-name",
				"internal_tag_name":     "some-internal-tag-name",
				"director_address":      "some-director-address",
				"concourse_target_pool": "some-concourse-target-pool",
				"concourse_lb_ip":       "some-concourse-lb-ip",
			}))
		})
	})

	Context("when tfState is empty", func() {
		BeforeEach(func() {
			executor.OutputCall.Stub = func(output string) (string, error) {
				return "", fmt.Errorf("unexpected output requested: %s", output)
			}
		})

		It("an empty map of outputs", func() {
			outputs, err := outputGenerator.Generate(storage.State{
				IAAS:  "gcp",
				EnvID: "some-env-id",
				GCP: storage.GCP{
					ServiceAccountKey: "some-service-account-key",
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "some-region",
				},
				TFState: "",
				LB: storage.LB{
					Type:   "",
					Domain: "",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(outputs).To(Equal(map[string]interface{}{}))
		})
	})

	Context("failure cases", func() {
		DescribeTable("returns an error when the outputter fails",
			func(outputName, lbType string) {
				expectedError := fmt.Sprintf("failed to get %s", outputName)
				executor.OutputCall.Stub = func(output string) (string, error) {
					if output == outputName {
						return "", errors.New(expectedError)
					}

					return "", nil
				}

				_, err := outputGenerator.Generate(storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
					TFState: "some-tf-state",
					LB: storage.LB{
						Type:   lbType,
						Domain: "some-domain",
					},
				})
				Expect(err).To(MatchError(expectedError))
			},
			Entry("failed to get external_ip", "external_ip", ""),
			Entry("failed to get network_name", "network_name", ""),
			Entry("failed to get subnetwork_name", "subnetwork_name", ""),
			Entry("failed to get bosh_open_tag_name", "bosh_open_tag_name", ""),
			Entry("failed to get internal_tag_name", "internal_tag_name", ""),
			Entry("failed to get director_address", "director_address", ""),
			Entry("failed to get router_backend_service", "router_backend_service", "cf"),
			Entry("failed to get ssh_proxy_target_pool", "ssh_proxy_target_pool", "cf"),
			Entry("failed to get tcp_router_target_pool", "tcp_router_target_pool", "cf"),
			Entry("failed to get ws_target_pool", "ws_target_pool", "cf"),
			Entry("failed to get router_lb_ip", "router_lb_ip", "cf"),
			Entry("failed to get ssh_proxy_lb_ip", "ssh_proxy_lb_ip", "cf"),
			Entry("failed to get tcp_router_lb_ip", "tcp_router_lb_ip", "cf"),
			Entry("failed to get ws_lb_ip", "ws_lb_ip", "cf"),
			Entry("failed to get system_domain_dns_servers", "system_domain_dns_servers", "cf"),
			Entry("failed to get concourse_target_pool", "concourse_target_pool", "concourse"),
			Entry("failed to get concourse_lb_ip", "concourse_lb_ip", "concourse"),
		)
	})
})
