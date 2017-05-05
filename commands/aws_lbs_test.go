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
		logger                *fakes.Logger

		incomingState storage.State
	)

	BeforeEach(func() {
		credentialValidator = &fakes.CredentialValidator{}
		infrastructureManager = &fakes.InfrastructureManager{}
		logger = &fakes.Logger{}

		command = commands.NewAWSLBs(credentialValidator, infrastructureManager, logger)
	})

	Describe("Execute", func() {
		Context("with cloudformation", func() {
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

				err := command.Execute(incomingState)
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

				err := command.Execute(incomingState)
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
				err := command.Execute(incomingState)

				Expect(err).To(MatchError("no lbs found"))
			})

			Context("failure cases", func() {
				Context("when credential validator fails", func() {
					It("returns an error", func() {
						credentialValidator.ValidateCall.Returns.Error = errors.New("validator failed")

						err := command.Execute(incomingState)

						Expect(err).To(MatchError("validator failed"))
					})
				})

				Context("when infrastructure manager fails", func() {
					It("returns an error", func() {
						infrastructureManager.DescribeCall.Returns.Error = errors.New("infrastructure manager failed")

						err := command.Execute(incomingState)

						Expect(err).To(MatchError("infrastructure manager failed"))
					})
				})
			})
		})
	})
})
