package commands_test

import (
	"bytes"
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LBs", func() {
	var (
		awsCredentialValidator *fakes.AWSCredentialValidator
		infrastructureManager  *fakes.InfrastructureManager
		stdout                 *bytes.Buffer
	)

	BeforeEach(func() {
		awsCredentialValidator = &fakes.AWSCredentialValidator{}
		infrastructureManager = &fakes.InfrastructureManager{}
		stdout = bytes.NewBuffer([]byte{})
	})

	Describe("Execute", func() {
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

			lbsCommand := commands.NewLBs(awsCredentialValidator, infrastructureManager, stdout)

			_, err := lbsCommand.Execute([]string{}, storage.State{
				Stack: storage.Stack{
					LBType: "cf",
					Name:   "some-stack-name",
				},
			})

			Expect(err).NotTo(HaveOccurred())

			Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(1))

			Expect(infrastructureManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))

			Expect(stdout.String()).To(ContainSubstring("some-lb-name: http://some.lb.url"))
			Expect(stdout.String()).To(ContainSubstring("some-other-lb-name: http://some.other.lb.url"))
		})

		It("prints LB names and URLs for lb type concourse", func() {
			infrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
				Name: "some-stack-name",
				Outputs: map[string]string{
					"ConcourseLoadBalancer":    "some-lb-name",
					"ConcourseLoadBalancerURL": "http://some.lb.url",
				},
			}

			lbsCommand := commands.NewLBs(awsCredentialValidator, infrastructureManager, stdout)

			_, err := lbsCommand.Execute([]string{}, storage.State{
				Stack: storage.Stack{
					LBType: "concourse",
					Name:   "some-stack-name",
				},
			})

			Expect(err).NotTo(HaveOccurred())

			Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(1))

			Expect(infrastructureManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))

			Expect(stdout.String()).To(ContainSubstring("some-lb-name: http://some.lb.url"))
		})

		It("returns error when lb type is not cf or concourse", func() {
			lbsCommand := commands.NewLBs(awsCredentialValidator, infrastructureManager, stdout)

			_, err := lbsCommand.Execute([]string{}, storage.State{
				Stack: storage.Stack{
					LBType: "",
				},
			})

			Expect(err).To(MatchError("no lbs found"))
		})

		Context("failure cases", func() {
			Context("when credential validator fails", func() {
				It("returns an error", func() {
					awsCredentialValidator.ValidateCall.Returns.Error = errors.New("validator failed")

					lbsCommand := commands.NewLBs(awsCredentialValidator, infrastructureManager, stdout)
					_, err := lbsCommand.Execute([]string{}, storage.State{})

					Expect(err).To(MatchError("validator failed"))
				})
			})

			Context("when infrastructure manager fails", func() {
				It("returns an error", func() {
					infrastructureManager.DescribeCall.Returns.Error = errors.New("infrastructure manager failed")

					lbsCommand := commands.NewLBs(awsCredentialValidator, infrastructureManager, stdout)
					_, err := lbsCommand.Execute([]string{}, storage.State{})

					Expect(err).To(MatchError("infrastructure manager failed"))
				})
			})
		})
	})
})
