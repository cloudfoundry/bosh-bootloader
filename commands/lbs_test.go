package commands_test

import (
	"bytes"
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
		terraformOutputter    *fakes.TerraformOutputter
		lbsCommand            commands.LBs
		stdout                *bytes.Buffer
		incomingState         storage.State
	)

	BeforeEach(func() {
		credentialValidator = &fakes.CredentialValidator{}
		infrastructureManager = &fakes.InfrastructureManager{}
		stateValidator = &fakes.StateValidator{}
		terraformOutputter = &fakes.TerraformOutputter{}
		stdout = bytes.NewBuffer([]byte{})

		lbsCommand = commands.NewLBs(credentialValidator, stateValidator, infrastructureManager, terraformOutputter, stdout)
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

				Expect(credentialValidator.ValidateAWSCall.CallCount).To(Equal(1))

				Expect(infrastructureManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))

				Expect(stdout.String()).To(ContainSubstring("CF Router LB: some-lb-name [http://some.lb.url]"))
				Expect(stdout.String()).To(ContainSubstring("CF SSH Proxy LB: some-other-lb-name [http://some.other.lb.url]"))
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

				Expect(credentialValidator.ValidateAWSCall.CallCount).To(Equal(1))

				Expect(infrastructureManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))

				Expect(stdout.String()).To(ContainSubstring("Concourse LB: some-lb-name [http://some.lb.url]"))
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
						credentialValidator.ValidateAWSCall.Returns.Error = errors.New("validator failed")

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
				terraformOutputter.GetCall.Stub = func(output string) (string, error) {
					switch output {
					case "router_lb_ip":
						return "some-router-lb-ip", nil
					case "ssh_proxy_lb_ip":
						return "some-ssh-proxy-lb-ip", nil
					case "tcp_router_lb_ip":
						return "some-tcp-router-lb-ip", nil
					case "concourse_lb_ip":
						return "some-concourse-lb-ip", nil
					case "ws_lb_ip":
						return "some-ws-lb-ip", nil
					default:
						return "", nil
					}
				}
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

				Expect(stdout.String()).To(ContainSubstring("CF Router LB: some-router-lb-ip"))
				Expect(stdout.String()).To(ContainSubstring("CF SSH Proxy LB: some-ssh-proxy-lb-ip"))
				Expect(stdout.String()).To(ContainSubstring("CF TCP Router LB: some-tcp-router-lb-ip"))
				Expect(stdout.String()).To(ContainSubstring("CF WebSocket LB: some-ws-lb-ip"))
			})

			It("prints LB ips for lb type concourse", func() {
				incomingState.LB = storage.LB{
					Type: "concourse",
				}
				err := lbsCommand.Execute([]string{}, incomingState)

				Expect(err).NotTo(HaveOccurred())

				Expect(stdout.String()).To(ContainSubstring("Concourse LB: some-concourse-lb-ip"))
			})

			Context("failure cases", func() {
				It("returns an error when terraform outputter fails to return router_lb_ip", func() {
					terraformOutputter.GetCall.Stub = func(output string) (string, error) {
						switch output {
						case "router_lb_ip":
							return "", errors.New("failed to return router_lb_ip")
						default:
							return "", nil
						}
					}
					incomingState.LB = storage.LB{
						Type: "cf",
					}
					err := lbsCommand.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("failed to return router_lb_ip"))
				})

				It("returns an error when terraform outputter fails to return ssh_proxy_lb_ip", func() {
					terraformOutputter.GetCall.Stub = func(output string) (string, error) {
						switch output {
						case "ssh_proxy_lb_ip":
							return "", errors.New("failed to return ssh_proxy_lb_ip")
						default:
							return "", nil
						}
					}
					incomingState.LB = storage.LB{
						Type: "cf",
					}
					err := lbsCommand.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("failed to return ssh_proxy_lb_ip"))
				})

				It("returns an error when terraform outputter fails to return ssh_proxy_lb_ip", func() {
					terraformOutputter.GetCall.Stub = func(output string) (string, error) {
						switch output {
						case "tcp_router_lb_ip":
							return "", errors.New("failed to return tcp_router_lb_ip")
						default:
							return "", nil
						}
					}
					incomingState.LB = storage.LB{
						Type: "cf",
					}
					err := lbsCommand.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("failed to return tcp_router_lb_ip"))
				})

				It("returns an error when terraform outputter fails to return ws_lb_ip", func() {
					terraformOutputter.GetCall.Stub = func(output string) (string, error) {
						switch output {
						case "ws_lb_ip":
							return "", errors.New("failed to return ws_lb_ip")
						default:
							return "", nil
						}
					}
					incomingState.LB = storage.LB{
						Type: "cf",
					}
					err := lbsCommand.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("failed to return ws_lb_ip"))
				})

				It("returns an error when terraform outputter fails to return concourse_lb_ip", func() {
					terraformOutputter.GetCall.Stub = func(output string) (string, error) {
						switch output {
						case "concourse_lb_ip":
							return "", errors.New("failed to return concourse_lb_ip")
						default:
							return "", nil
						}
					}
					incomingState.LB = storage.LB{
						Type: "concourse",
					}
					err := lbsCommand.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("failed to return concourse_lb_ip"))
				})

				It("returns an error when terraform outputter fails to return concourse_lb_ip", func() {
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
