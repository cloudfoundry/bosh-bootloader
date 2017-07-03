package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWSLBs", func() {
	var (
		command commands.AWSLBs

		credentialValidator   *fakes.CredentialValidator
		infrastructureManager *fakes.InfrastructureManager
		terraformManager      *fakes.TerraformManager
		logger                *fakes.Logger

		incomingState storage.State
	)

	BeforeEach(func() {
		credentialValidator = &fakes.CredentialValidator{}
		infrastructureManager = &fakes.InfrastructureManager{}
		terraformManager = &fakes.TerraformManager{}
		logger = &fakes.Logger{}

		command = commands.NewAWSLBs(credentialValidator, infrastructureManager, terraformManager, logger)
	})

	Describe("Execute", func() {
		Context("with cloudformation", func() {
			BeforeEach(func() {
				incomingState.TFState = ""
			})
			It("prints LB names and URLs for lb type cf", func() {
				infrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
					Name: "some-stack-name",
					Outputs: map[string]string{
						"CFRouterLoadBalancer":      "some-lb-name",
						"CFRouterLoadBalancerURL":   "http://some.lb.url",
						"CFSSHProxyLoadBalancer":    "some-other-lb-name",
						"CFSSHProxyLoadBalancerURL": "http://some.other.lb.url",
					},
				}
				incomingState.Stack = storage.Stack{
					LBType: "cf",
					Name:   "some-stack-name",
				}

				err := command.Execute([]string{}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(credentialValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(infrastructureManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))

				Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
					"CF Router LB: some-lb-name [http://some.lb.url]\n",
					"CF SSH Proxy LB: some-other-lb-name [http://some.other.lb.url]\n",
				}))
			})

			It("prints LB names and URLs for lb type concourse", func() {
				infrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
					Name: "some-stack-name",
					Outputs: map[string]string{
						"ConcourseLoadBalancer":    "some-lb-name",
						"ConcourseLoadBalancerURL": "http://some.lb.url",
					},
				}
				incomingState.Stack = storage.Stack{
					LBType: "concourse",
					Name:   "some-stack-name",
				}

				err := command.Execute([]string{}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(credentialValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(infrastructureManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))

				Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
					"Concourse LB: some-lb-name [http://some.lb.url]\n",
				}))
			})

			It("returns error when lb type is not cf or concourse", func() {
				incomingState.Stack = storage.Stack{
					LBType: "",
				}
				err := command.Execute([]string{}, incomingState)

				Expect(err).To(MatchError("no lbs found"))
			})

			Context("failure cases", func() {
				Context("when credential validator fails", func() {
					It("returns an error", func() {
						credentialValidator.ValidateCall.Returns.Error = errors.New("validator failed")

						err := command.Execute([]string{}, incomingState)

						Expect(err).To(MatchError("validator failed"))
					})
				})

				Context("when infrastructure manager fails", func() {
					It("returns an error", func() {
						infrastructureManager.DescribeCall.Returns.Error = errors.New("infrastructure manager failed")

						err := command.Execute([]string{}, incomingState)

						Expect(err).To(MatchError("infrastructure manager failed"))
					})
				})
			})
		})

		Context("with terraform", func() {
			Context("when the lb type is cf", func() {
				BeforeEach(func() {
					incomingState = storage.State{
						IAAS:    "aws",
						TFState: "some-tf-state",
						LB: storage.LB{
							Type: "cf",
						},
					}
					terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
						"cf_router_load_balancer":         "some-router-lb-name",
						"cf_router_load_balancer_url":     "some-router-lb-url",
						"cf_ssh_proxy_load_balancer":      "some-ssh-proxy-lb-name",
						"cf_ssh_proxy_load_balancer_url":  "some-ssh-proxy-lb-url",
						"cf_tcp_router_load_balancer":     "some-tcp-router-lb-name",
						"cf_tcp_router_load_balancer_url": "some-tcp-router-lb-url",
					}
				})

				It("prints LB names and URLs for router and ssh proxy", func() {
					err := command.Execute([]string{}, incomingState)

					Expect(err).NotTo(HaveOccurred())

					Expect(credentialValidator.ValidateCall.CallCount).To(Equal(1))
					Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
						"CF Router LB: some-router-lb-name [some-router-lb-url]\n",
						"CF SSH Proxy LB: some-ssh-proxy-lb-name [some-ssh-proxy-lb-url]\n",
						"CF TCP Router LB: some-tcp-router-lb-name [some-tcp-router-lb-url]\n",
					}))
				})

				Context("when the domain is specified", func() {
					BeforeEach(func() {
						incomingState.LB.Domain = "some-domain"

						terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
							"cf_router_load_balancer":         "some-router-lb-name",
							"cf_router_load_balancer_url":     "some-router-lb-url",
							"cf_ssh_proxy_load_balancer":      "some-ssh-proxy-lb-name",
							"cf_ssh_proxy_load_balancer_url":  "some-ssh-proxy-lb-url",
							"cf_tcp_router_load_balancer":     "some-tcp-router-lb-name",
							"cf_tcp_router_load_balancer_url": "some-tcp-router-lb-url",
							"cf_system_domain_dns_servers":    []string{"name-server-1.", "name-server-2."},
						}
					})

					It("prints LB names, URLs, and DNS servers", func() {
						err := command.Execute([]string{}, incomingState)

						Expect(err).NotTo(HaveOccurred())

						Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
							"CF Router LB: some-router-lb-name [some-router-lb-url]\n",
							"CF SSH Proxy LB: some-ssh-proxy-lb-name [some-ssh-proxy-lb-url]\n",
							"CF TCP Router LB: some-tcp-router-lb-name [some-tcp-router-lb-url]\n",
							"CF System Domain DNS servers: name-server-1. name-server-2.\n",
						}))
					})

					Context("when the json flag is provided", func() {
						It("prints LB names, URLs, and DNS servers in json format", func() {
							incomingState.LB = storage.LB{
								Type:   "cf",
								Domain: "some-domain",
							}
							err := command.Execute([]string{"--json"}, incomingState)
							Expect(err).NotTo(HaveOccurred())

							Expect(logger.PrintlnCall.Receives.Message).To(MatchJSON(`{
								"cf_router_lb": "some-router-lb-name",
								"cf_router_lb_url": "some-router-lb-url",
								"cf_ssh_proxy_lb": "some-ssh-proxy-lb-name",
								"cf_ssh_proxy_lb_url": "some-ssh-proxy-lb-url",
								"cf_tcp_lb": "some-tcp-router-lb-name",
								"cf_tcp_lb_url":  "some-tcp-router-lb-url",
								"env_dns_zone_name_servers": [
									"name-server-1.",
									"name-server-2."
								]
							}`))
						})
					})
				})
			})

			Context("when the lb type is concourse", func() {
				BeforeEach(func() {
					incomingState = storage.State{
						IAAS:    "aws",
						TFState: "some-tf-state",
						LB: storage.LB{
							Type: "concourse",
						},
					}
					terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
						"concourse_load_balancer":     "some-concourse-lb-name",
						"concourse_load_balancer_url": "some-concourse-lb-url",
					}
				})

				It("prints LB name and URL", func() {
					err := command.Execute([]string{}, incomingState)

					Expect(err).NotTo(HaveOccurred())

					Expect(credentialValidator.ValidateCall.CallCount).To(Equal(1))
					Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
						"Concourse LB: some-concourse-lb-name [some-concourse-lb-url]\n",
					}))
				})
			})

			It("returns error when lb type is not cf or concourse", func() {
				incomingState = storage.State{
					IAAS:    "aws",
					TFState: "some-tf-state",
					LB: storage.LB{
						Type: "other",
					},
				}
				err := command.Execute([]string{}, incomingState)

				Expect(err).To(MatchError("no lbs found"))
			})

			Context("failure cases", func() {
				BeforeEach(func() {
					incomingState = storage.State{
						TFState: "some-tf-state",
					}
				})

				Context("when credential validator fails", func() {
					It("returns an error", func() {
						credentialValidator.ValidateCall.Returns.Error = errors.New("validator failed")

						err := command.Execute([]string{}, incomingState)

						Expect(err).To(MatchError("validator failed"))
					})
				})

				Context("when terraform manager fails", func() {
					It("returns an error", func() {
						terraformManager.GetOutputsCall.Returns.Error = errors.New("terraform manager failed")

						err := command.Execute([]string{}, incomingState)

						Expect(err).To(MatchError("terraform manager failed"))
					})
				})
			})
		})
	})
})
