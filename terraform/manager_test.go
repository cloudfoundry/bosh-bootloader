package terraform_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("Manager", func() {
	var (
		executor             *fakes.TerraformExecutor
		gcpTemplateGenerator *fakes.GCPTemplateGenerator
		gcpInputGenerator    *fakes.GCPInputGenerator
		logger               *fakes.Logger
		manager              terraform.Manager
	)

	BeforeEach(func() {
		executor = &fakes.TerraformExecutor{}
		gcpTemplateGenerator = &fakes.GCPTemplateGenerator{}
		gcpInputGenerator = &fakes.GCPInputGenerator{}
		logger = &fakes.Logger{}

		manager = terraform.NewManager(executor, gcpTemplateGenerator, gcpInputGenerator, logger)
	})

	Describe("Apply", func() {
		var (
			incomingState   storage.State
			expectedState   storage.State
			expectedTFState string
		)

		BeforeEach(func() {
			incomingState = storage.State{
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
			}

			expectedTFState = "some-updated-tf-state"
			executor.ApplyCall.Returns.TFState = expectedTFState

			expectedState = incomingState
			expectedState.TFState = expectedTFState

			gcpTemplateGenerator.GenerateCall.Returns.Template = "some-gcp-terraform-template"
			gcpInputGenerator.GenerateCall.Returns.Inputs = map[string]string{
				"env_id":        incomingState.EnvID,
				"project_id":    incomingState.GCP.ProjectID,
				"region":        incomingState.GCP.Region,
				"zone":          incomingState.GCP.Zone,
				"credentials":   "some-path",
				"system_domain": incomingState.LB.Domain,
			}
		})

		It("returns a state with new tfState from executor apply", func() {
			_, err := manager.Apply(storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Messages).To(ContainSequence([]string{
				"generating terraform template", "applied terraform template",
			}))
		})

		It("returns a state with new tfState from executor apply", func() {
			state, err := manager.Apply(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(gcpTemplateGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			Expect(gcpInputGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			Expect(executor.ApplyCall.Receives.Inputs).To(Equal(map[string]string{
				"env_id":        incomingState.EnvID,
				"project_id":    incomingState.GCP.ProjectID,
				"region":        incomingState.GCP.Region,
				"zone":          incomingState.GCP.Zone,
				"credentials":   "some-path",
				"system_domain": incomingState.LB.Domain,
			}))
			Expect(executor.ApplyCall.Receives.TFState).To(Equal("some-tf-state"))
			Expect(executor.ApplyCall.Receives.Template).To(Equal(string("some-gcp-terraform-template")))
			Expect(state).To(Equal(expectedState))
		})

		Context("failure cases", func() {
			Context("when InputGenerator.Generate returns an error", func() {
				BeforeEach(func() {
					gcpInputGenerator.GenerateCall.Returns.Error = errors.New("failed to generate inputs")
				})

				It("bubbles up the error", func() {
					_, err := manager.Apply(incomingState)
					Expect(err).To(MatchError("failed to generate inputs"))
				})
			})

			Context("when Executor.Apply returns a ExecutorError", func() {
				var (
					tempDir       string
					executorError *fakes.TerraformExecutorError
				)

				BeforeEach(func() {
					var err error
					tempDir, err = ioutil.TempDir("", "")
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(filepath.Join(tempDir, "terraform.tfstate"), []byte("updated-tf-state"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					executorError = &fakes.TerraformExecutorError{}
					executor.ApplyCall.Returns.Error = executorError
				})

				AfterEach(func() {
					executor.ApplyCall.Returns.Error = nil
				})

				It("returns a ManagerError", func() {
					_, err := manager.Apply(incomingState)

					expectedError := terraform.NewManagerError(incomingState, executorError)
					Expect(err).To(MatchError(expectedError))
				})
			})

			Context("when Executor.Apply returns a non-ExecutorError error", func() {
				executorError := errors.New("some-error")

				BeforeEach(func() {
					executor.ApplyCall.Returns.Error = executorError
				})

				AfterEach(func() {
					executor.ApplyCall.Returns.Error = nil
				})

				It("bubbles up the error", func() {
					_, err := manager.Apply(incomingState)
					Expect(err).To(Equal(executorError))
				})
			})
		})
	})

	Describe("Destroy", func() {
		Context("when the bbl state contains a non-empty TFState", func() {
			var (
				originalBBLState = storage.State{
					EnvID: "some-env-id",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
					LB: storage.LB{
						Type:   "cf",
						Domain: "some-domain",
					},
					TFState: "some-tf-state",
				}
				updatedTFState = "some-updated-tf-state"
			)

			BeforeEach(func() {
				gcpTemplateGenerator.GenerateCall.Returns.Template = "some-gcp-terraform-template"
			})

			It("calls Executor.Destroy with the right arguments", func() {
				_, err := manager.Destroy(originalBBLState)
				Expect(err).NotTo(HaveOccurred())

				Expect(gcpTemplateGenerator.GenerateCall.Receives.State).To(Equal(originalBBLState))

				Expect(executor.DestroyCall.Receives.Credentials).To(Equal(originalBBLState.GCP.ServiceAccountKey))
				Expect(executor.DestroyCall.Receives.EnvID).To(Equal(originalBBLState.EnvID))
				Expect(executor.DestroyCall.Receives.ProjectID).To(Equal(originalBBLState.GCP.ProjectID))
				Expect(executor.DestroyCall.Receives.Zone).To(Equal(originalBBLState.GCP.Zone))
				Expect(executor.DestroyCall.Receives.Region).To(Equal(originalBBLState.GCP.Region))
				Expect(executor.DestroyCall.Receives.Template).To(Equal(gcpTemplateGenerator.GenerateCall.Returns.Template))
				Expect(executor.DestroyCall.Receives.TFState).To(Equal(originalBBLState.TFState))
			})

			Context("when Executor.Destroy succeeds", func() {
				BeforeEach(func() {
					executor.DestroyCall.Returns.TFState = updatedTFState
				})

				AfterEach(func() {
					executor.DestroyCall.Returns.TFState = ""
				})

				It("returns the bbl state updated with the TFState returned by Executor.Destroy", func() {
					newBBLState, err := manager.Destroy(originalBBLState)
					Expect(err).NotTo(HaveOccurred())

					expectedBBLState := originalBBLState
					expectedBBLState.TFState = updatedTFState
					Expect(newBBLState.TFState).To(Equal(updatedTFState))
					Expect(newBBLState).To(Equal(expectedBBLState))
				})
			})

			Context("when Executor.Destroy returns a ExecutorError", func() {
				var (
					tempDir       string
					executorError *fakes.TerraformExecutorError
				)

				BeforeEach(func() {
					var err error
					tempDir, err = ioutil.TempDir("", "")
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(filepath.Join(tempDir, "terraform.tfstate"), []byte("updated-tf-state"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					executorError = &fakes.TerraformExecutorError{}
					executor.DestroyCall.Returns.Error = executorError
				})

				AfterEach(func() {
					executor.DestroyCall.Returns.Error = nil
				})

				It("returns a ManagerError", func() {
					_, err := manager.Destroy(originalBBLState)

					expectedError := terraform.NewManagerError(originalBBLState, executorError)
					Expect(err).To(MatchError(expectedError))
				})
			})

			Context("when Executor.Destroy returns a non-ExecutorError error", func() {
				executorError := errors.New("some-error")

				BeforeEach(func() {
					executor.DestroyCall.Returns.Error = executorError
				})

				AfterEach(func() {
					executor.DestroyCall.Returns.Error = nil
				})

				It("bubbles up the error", func() {
					_, err := manager.Destroy(originalBBLState)
					Expect(err).To(Equal(executorError))
				})
			})
		})

		Context("when the bbl state contains a non-empty TFState", func() {
			var (
				originalBBLState = storage.State{
					EnvID: "some-env-id",
				}
			)

			It("returns the bbl state and skips calling executor destroy", func() {
				bblState, err := manager.Destroy(originalBBLState)
				Expect(err).NotTo(HaveOccurred())

				Expect(bblState).To(Equal(originalBBLState))
				Expect(executor.DestroyCall.CallCount).To(Equal(0))
			})
		})
	})

	Describe("GetOutputs", func() {
		BeforeEach(func() {
			executor.OutputCall.Stub = func(output string) (string, error) {
				switch output {
				case "network_name":
					return "some-network-name", nil
				case "subnetwork_name":
					return "some-subnetwork-name", nil
				case "bosh_open_tag_name":
					return "some-bosh-open-tag-name", nil
				case "internal_tag_name":
					return "some-internal-tag-name", nil
				case "external_ip":
					return "some-external-ip", nil
				case "director_address":
					return "some-director-address", nil
				case "router_backend_service":
					return "some-router-backend-service", nil
				case "ws_lb_ip":
					return "some-ws-lb-ip", nil
				case "ssh_proxy_lb_ip":
					return "some-ssh-proxy-lb-ip", nil
				case "tcp_router_lb_ip":
					return "some-tcp-router-lb-ip", nil
				case "router_lb_ip":
					return "some-router-lb-ip", nil
				case "ssh_proxy_target_pool":
					return "some-ssh-proxy-target-pool", nil
				case "tcp_router_target_pool":
					return "some-tcp-router-target-pool", nil
				case "ws_target_pool":
					return "some-ws-target-pool", nil
				case "concourse_target_pool":
					return "some-concourse-target-pool", nil
				case "concourse_lb_ip":
					return "some-concourse-lb-ip", nil
				case "system_domain_dns_servers":
					return "", errors.New("no dns server exists")
				default:
					return "", nil
				}
			}
		})

		Context("when no lb exists", func() {
			It("returns all terraform outputs except lb related outputs", func() {
				terraformOutputs, err := manager.GetOutputs("some-tf-state", "", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(executor.OutputCall.Receives.TFState).To(Equal("some-tf-state"))
				Expect(terraformOutputs).To(Equal(terraform.Outputs{
					ExternalIP:      "some-external-ip",
					NetworkName:     "some-network-name",
					SubnetworkName:  "some-subnetwork-name",
					BOSHTag:         "some-bosh-open-tag-name",
					InternalTag:     "some-internal-tag-name",
					DirectorAddress: "some-director-address",
				}))
			})
		})

		Context("when cf lb exists", func() {
			Context("when the domain is not specified", func() {
				It("returns terraform outputs related to cf lb without system domain DNS servers", func() {
					terraformOutputs, err := manager.GetOutputs("some-tf-state", "cf", false)
					Expect(err).NotTo(HaveOccurred())
					Expect(executor.OutputCall.Receives.TFState).To(Equal("some-tf-state"))
					Expect(terraformOutputs).To(Equal(terraform.Outputs{
						ExternalIP:           "some-external-ip",
						NetworkName:          "some-network-name",
						SubnetworkName:       "some-subnetwork-name",
						BOSHTag:              "some-bosh-open-tag-name",
						InternalTag:          "some-internal-tag-name",
						DirectorAddress:      "some-director-address",
						RouterBackendService: "some-router-backend-service",
						SSHProxyTargetPool:   "some-ssh-proxy-target-pool",
						TCPRouterTargetPool:  "some-tcp-router-target-pool",
						WSTargetPool:         "some-ws-target-pool",
						RouterLBIP:           "some-router-lb-ip",
						SSHProxyLBIP:         "some-ssh-proxy-lb-ip",
						TCPRouterLBIP:        "some-tcp-router-lb-ip",
						WebSocketLBIP:        "some-ws-lb-ip",
					}))
				})
			})

			Context("when the domain is specified", func() {
				It("returns terraform outputs related to cf lb with the system domain DNS servers", func() {
					executor.OutputCall.Stub = func(output string) (string, error) {
						switch output {
						case "network_name":
							return "some-network-name", nil
						case "subnetwork_name":
							return "some-subnetwork-name", nil
						case "bosh_open_tag_name":
							return "some-bosh-open-tag-name", nil
						case "internal_tag_name":
							return "some-internal-tag-name", nil
						case "external_ip":
							return "some-external-ip", nil
						case "director_address":
							return "some-director-address", nil
						case "router_backend_service":
							return "some-router-backend-service", nil
						case "ws_lb_ip":
							return "some-ws-lb-ip", nil
						case "ssh_proxy_lb_ip":
							return "some-ssh-proxy-lb-ip", nil
						case "tcp_router_lb_ip":
							return "some-tcp-router-lb-ip", nil
						case "router_lb_ip":
							return "some-router-lb-ip", nil
						case "ssh_proxy_target_pool":
							return "some-ssh-proxy-target-pool", nil
						case "tcp_router_target_pool":
							return "some-tcp-router-target-pool", nil
						case "ws_target_pool":
							return "some-ws-target-pool", nil
						case "concourse_target_pool":
							return "some-concourse-target-pool", nil
						case "concourse_lb_ip":
							return "some-concourse-lb-ip", nil
						case "system_domain_dns_servers":
							return "some-name-server-1,\nsome-name-server-2", nil
						default:
							return "", nil
						}
					}

					terraformOutputs, err := manager.GetOutputs("some-tf-state", "cf", true)
					Expect(err).NotTo(HaveOccurred())
					Expect(executor.OutputCall.Receives.TFState).To(Equal("some-tf-state"))
					Expect(terraformOutputs).To(Equal(terraform.Outputs{
						ExternalIP:             "some-external-ip",
						NetworkName:            "some-network-name",
						SubnetworkName:         "some-subnetwork-name",
						BOSHTag:                "some-bosh-open-tag-name",
						InternalTag:            "some-internal-tag-name",
						DirectorAddress:        "some-director-address",
						RouterBackendService:   "some-router-backend-service",
						SSHProxyTargetPool:     "some-ssh-proxy-target-pool",
						TCPRouterTargetPool:    "some-tcp-router-target-pool",
						WSTargetPool:           "some-ws-target-pool",
						RouterLBIP:             "some-router-lb-ip",
						SSHProxyLBIP:           "some-ssh-proxy-lb-ip",
						TCPRouterLBIP:          "some-tcp-router-lb-ip",
						WebSocketLBIP:          "some-ws-lb-ip",
						SystemDomainDNSServers: []string{"some-name-server-1", "some-name-server-2"},
					}))
				})
			})
		})

		Context("when concourse lb exists", func() {
			It("returns terraform outputs related to concourse lb", func() {
				terraformOutputs, err := manager.GetOutputs("some-tf-state", "concourse", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(executor.OutputCall.Receives.TFState).To(Equal("some-tf-state"))
				Expect(terraformOutputs).To(Equal(terraform.Outputs{
					ExternalIP:          "some-external-ip",
					NetworkName:         "some-network-name",
					SubnetworkName:      "some-subnetwork-name",
					BOSHTag:             "some-bosh-open-tag-name",
					InternalTag:         "some-internal-tag-name",
					DirectorAddress:     "some-director-address",
					ConcourseTargetPool: "some-concourse-target-pool",
					ConcourseLBIP:       "some-concourse-lb-ip",
				}))
			})
		})

		Context("when tfState is empty", func() {
			It("returns an empty terraform outputs", func() {
				terraformOutputs, err := manager.GetOutputs("", "concourse", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(executor.OutputCall.CallCount).To(Equal(0))
				Expect(terraformOutputs).To(Equal(terraform.Outputs{}))
			})
		})

		Context("failure cases", func() {
			DescribeTable("returns an error when the outputter fails", func(outputName, lbType string) {
				expectedError := fmt.Sprintf("failed to get %s", outputName)
				executor.OutputCall.Stub = func(output string) (string, error) {
					if output == outputName {
						return "", errors.New(expectedError)
					}
					return "", nil
				}

				_, err := manager.GetOutputs(outputName, lbType, true)
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
				Entry("failed to get concourse_target_pool", "concourse_target_pool", "concourse"),
				Entry("failed to get concourse_lb_ip", "concourse_lb_ip", "concourse"),
				Entry("failed to get system_domain_dns_servers", "system_domain_dns_servers", "cf"),
			)
		})
	})

	Describe("Version", func() {
		BeforeEach(func() {
			executor.VersionCall.Returns.Version = "some-version"
		})

		It("returns a version", func() {
			version, err := manager.Version()
			Expect(err).NotTo(HaveOccurred())

			Expect(executor.VersionCall.CallCount).To(Equal(1))
			Expect(version).To(Equal("some-version"))
		})

		Context("failure cases", func() {
			Context("when version returns an error", func() {
				BeforeEach(func() {
					executor.VersionCall.Returns.Error = errors.New("failed to get version")
				})

				It("returns the error", func() {
					_, err := manager.Version()
					Expect(err).To(MatchError("failed to get version"))
				})
			})
		})
	})

	Describe("ValidateVersion", func() {
		Context("when terraform version is greater than v0.8.5", func() {
			BeforeEach(func() {
				executor.VersionCall.Returns.Version = "0.9.1"
			})

			It("validates the version of terraform and returns no error", func() {
				err := manager.ValidateVersion()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when terraform version is v0.9.0", func() {
			BeforeEach(func() {
				executor.VersionCall.Returns.Version = "0.9.0"
			})

			It("returns a helpful error message", func() {
				err := manager.ValidateVersion()
				Expect(err).To(MatchError("Version 0.9.0 of terraform is incompatible with bbl, please try a later version."))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the terraform installed is less than v0.8.5", func() {
				executor.VersionCall.Returns.Version = "0.8.4"

				err := manager.ValidateVersion()
				Expect(err).To(MatchError("Terraform version must be at least v0.8.5"))
			})

			It("fast fails if the terraform executor fails to get the version", func() {
				executor.VersionCall.Returns.Error = errors.New("cannot get version")

				err := manager.ValidateVersion()
				Expect(err).To(MatchError("cannot get version"))
			})

			It("fast fails when the version cannot be parsed by go-semver", func() {
				executor.VersionCall.Returns.Version = "lol.5.2"

				err := manager.ValidateVersion()
				Expect(err.Error()).To(ContainSubstring("invalid syntax"))
			})
		})
	})
})
