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

var _ = Describe("LBs", func() {
	var (
		credentialValidator   *fakes.CredentialValidator
		infrastructureManager *fakes.InfrastructureManager
		stateValidator        *fakes.StateValidator
		terraformManager      *fakes.TerraformManager
		lbsCommand            commands.LBs
		logger                *fakes.Logger
		incomingState         storage.State
	)

	BeforeEach(func() {
		credentialValidator = &fakes.CredentialValidator{}
		infrastructureManager = &fakes.InfrastructureManager{}
		stateValidator = &fakes.StateValidator{}
		terraformManager = &fakes.TerraformManager{}
		terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
			"router_lb_ip":     "some-router-lb-ip",
			"ssh_proxy_lb_ip":  "some-ssh-proxy-lb-ip",
			"tcp_router_lb_ip": "some-tcp-router-lb-ip",
			"ws_lb_ip":         "some-ws-lb-ip",
			"concourse_lb_ip":  "some-concourse-lb-ip",
		}
		logger = &fakes.Logger{}

		lbsCommand = commands.NewLBs(credentialValidator, stateValidator, infrastructureManager, terraformManager, logger)
	})

	Describe("Execute", func() {
		Context("when bbl'd up on aws", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					IAAS: "aws",
				}
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

				err := lbsCommand.Execute([]string{}, incomingState)
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

				err := lbsCommand.Execute([]string{}, incomingState)
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
				err := lbsCommand.Execute([]string{}, incomingState)

				Expect(err).To(MatchError("no lbs found"))
			})

			Context("failure cases", func() {
				Context("when credential validator fails", func() {
					It("returns an error", func() {
						credentialValidator.ValidateCall.Returns.Error = errors.New("validator failed")

						err := lbsCommand.Execute([]string{}, incomingState)

						Expect(err).To(MatchError("validator failed"))
					})
				})

				Context("when infrastructure manager fails", func() {
					It("returns an error", func() {
						infrastructureManager.DescribeCall.Returns.Error = errors.New("infrastructure manager failed")

						err := lbsCommand.Execute([]string{}, incomingState)

						Expect(err).To(MatchError("infrastructure manager failed"))
					})
				})
			})
		})

		Context("when bbl'd up on gcp", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					IAAS: "gcp",
				}
			})

			It("prints LB ips for lb type cf", func() {
				incomingState.LB = storage.LB{
					Type: "cf",
				}
				err := lbsCommand.Execute([]string{}, incomingState)

				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
					"CF Router LB: some-router-lb-ip\n",
					"CF SSH Proxy LB: some-ssh-proxy-lb-ip\n",
					"CF TCP Router LB: some-tcp-router-lb-ip\n",
					"CF WebSocket LB: some-ws-lb-ip\n",
				}))
			})

			Context("when the domain is specified", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
						"router_lb_ip":              "some-router-lb-ip",
						"ssh_proxy_lb_ip":           "some-ssh-proxy-lb-ip",
						"tcp_router_lb_ip":          "some-tcp-router-lb-ip",
						"ws_lb_ip":                  "some-ws-lb-ip",
						"concourse_lb_ip":           "some-concourse-lb-ip",
						"system_domain_dns_servers": []string{"name-server-1.", "name-server-2."},
					}
				})

				It("prints LB ips for lb type cf in human readable format", func() {
					incomingState.LB = storage.LB{
						Type:   "cf",
						Domain: "some-domain",
					}
					err := lbsCommand.Execute([]string{}, incomingState)

					Expect(err).NotTo(HaveOccurred())

					Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
						"CF Router LB: some-router-lb-ip\n",
						"CF SSH Proxy LB: some-ssh-proxy-lb-ip\n",
						"CF TCP Router LB: some-tcp-router-lb-ip\n",
						"CF WebSocket LB: some-ws-lb-ip\n",
						"CF System Domain DNS servers: name-server-1. name-server-2.\n",
					}))
				})

				Context("when the json flag is provided", func() {
					It("prints LB ips for lb type cf in json format", func() {
						incomingState.LB = storage.LB{
							Type:   "cf",
							Domain: "some-domain",
						}
						err := lbsCommand.Execute([]string{"--json"}, incomingState)
						Expect(err).NotTo(HaveOccurred())

						Expect(logger.PrintlnCall.Receives.Message).To(MatchJSON(`{
							"cf_router_lb": "some-router-lb-ip",
							"cf_ssh_proxy_lb": "some-ssh-proxy-lb-ip",
							"cf_tcp_router_lb": "some-tcp-router-lb-ip",
							"cf_websocket_lb": "some-ws-lb-ip",
							"cf_system_domain_dns_servers": [
								"name-server-1.",
								"name-server-2."
							]
						}`))
					})
				})
			})

			It("prints LB ips for lb type concourse", func() {
				incomingState.LB = storage.LB{
					Type: "concourse",
				}
				err := lbsCommand.Execute([]string{}, incomingState)

				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
					"Concourse LB: some-concourse-lb-ip\n",
				}))
			})

			Context("failure cases", func() {
				It("returns an error when terraform output provider fails", func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to return terraform output")
					err := lbsCommand.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("failed to return terraform output"))
				})

				It("returns an nice error message when no lb type is found", func() {
					incomingState.LB = storage.LB{
						Type: "",
					}
					err := lbsCommand.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("no lbs found"))
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when state validator fails", func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")

				err := lbsCommand.Execute([]string{}, storage.State{})

				Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(err).To(MatchError("state validator failed"))
			})
		})
	})
})
