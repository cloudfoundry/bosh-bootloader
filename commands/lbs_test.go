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
		lbsCommand commands.LBs

		gcpLBs                *fakes.GCPLBs
		credentialValidator   *fakes.CredentialValidator
		infrastructureManager *fakes.InfrastructureManager
		stateValidator        *fakes.StateValidator
		logger                *fakes.Logger

		incomingState storage.State
	)

	BeforeEach(func() {
		gcpLBs = &fakes.GCPLBs{}

		credentialValidator = &fakes.CredentialValidator{}
		infrastructureManager = &fakes.InfrastructureManager{}
		stateValidator = &fakes.StateValidator{}
		logger = &fakes.Logger{}

		lbsCommand = commands.NewLBs(gcpLBs, credentialValidator, stateValidator, infrastructureManager, logger)
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
			It("prints LB ips for lb type cf", func() {
				incomingState := storage.State{
					IAAS: "gcp",
				}
				err := lbsCommand.Execute([]string{}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(gcpLBs.ExecuteCall.Receives.SubcommandFlags).To(Equal([]string{}))
				Expect(gcpLBs.ExecuteCall.Receives.State).To(Equal(incomingState))
			})
		})

		Context("failure cases", func() {
			It("returns an error when state validator fails", func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")

				err := lbsCommand.Execute([]string{}, storage.State{})

				Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(err).To(MatchError("state validator failed"))
			})

			It("returns an error when the GCPLBs fails", func() {
				gcpLBs.ExecuteCall.Returns.Error = errors.New("something bad happened")

				err := lbsCommand.Execute([]string{}, storage.State{
					IAAS: "gcp",
				})
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})
})
