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
		awsCredentialValidator *fakes.AWSCredentialValidator
		infrastructureManager  *fakes.InfrastructureManager
		stateValidator         *fakes.StateValidator
		lbsCommand             commands.LBs
		stdout                 *bytes.Buffer
	)

	BeforeEach(func() {
		awsCredentialValidator = &fakes.AWSCredentialValidator{}
		infrastructureManager = &fakes.InfrastructureManager{}
		stateValidator = &fakes.StateValidator{}
		stdout = bytes.NewBuffer([]byte{})

		lbsCommand = commands.NewLBs(awsCredentialValidator, stateValidator, infrastructureManager, stdout)
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

			err := lbsCommand.Execute([]string{}, storage.State{
				Stack: storage.Stack{
					LBType: "cf",
					Name:   "some-stack-name",
				},
			})

			Expect(err).NotTo(HaveOccurred())

			Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(1))

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

			err := lbsCommand.Execute([]string{}, storage.State{
				Stack: storage.Stack{
					LBType: "concourse",
					Name:   "some-stack-name",
				},
			})

			Expect(err).NotTo(HaveOccurred())

			Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(1))

			Expect(infrastructureManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))

			Expect(stdout.String()).To(ContainSubstring("Concourse LB: some-lb-name [http://some.lb.url]"))
		})

		It("returns error when lb type is not cf or concourse", func() {
			err := lbsCommand.Execute([]string{}, storage.State{
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

					err := lbsCommand.Execute([]string{}, storage.State{})

					Expect(err).To(MatchError("validator failed"))
				})
			})

			Context("when infrastructure manager fails", func() {
				It("returns an error", func() {
					infrastructureManager.DescribeCall.Returns.Error = errors.New("infrastructure manager failed")

					err := lbsCommand.Execute([]string{}, storage.State{})

					Expect(err).To(MatchError("infrastructure manager failed"))
				})
			})

			It("returns an error when state validator fails", func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")

				err := lbsCommand.Execute([]string{}, storage.State{})

				Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(err).To(MatchError("state validator failed"))
			})

		})
	})
})
