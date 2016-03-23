package unsupported_test

import (
	"errors"
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InfrastructureCreator", func() {
	var (
		builder              *fakes.TemplateBuilder
		stackManager         *fakes.StackManager
		cloudFormationClient *fakes.CloudFormationClient
		creator              unsupported.InfrastructureCreator
	)

	BeforeEach(func() {
		cloudFormationClient = &fakes.CloudFormationClient{}

		builder = &fakes.TemplateBuilder{}
		builder.BuildCall.Returns.Template = templates.Template{
			AWSTemplateFormatVersion: "some-template-version",
			Description:              "some-description",
		}

		stackManager = &fakes.StackManager{}
		stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{
			Name: "concourse",
			Outputs: map[string]string{
				"BOSHSubnet": "some-subnet-id",
			},
		}

		creator = unsupported.NewInfrastructureCreator(builder, stackManager)
	})

	It("creates the underlying infrastructure and returns the stack", func() {
		stack, err := creator.Create("some-key-pair-name", 2, cloudFormationClient)
		Expect(err).NotTo(HaveOccurred())

		Expect(stack).To(Equal(cloudformation.Stack{
			Name: "concourse",
			Outputs: map[string]string{
				"BOSHSubnet": "some-subnet-id",
			},
		}))
		Expect(builder.BuildCall.Receives.KeyPairName).To(Equal("some-key-pair-name"))
		Expect(builder.BuildCall.Receives.NumberOfAZs).To(Equal(2))

		Expect(stackManager.CreateOrUpdateCall.Receives.Client).To(Equal(cloudFormationClient))
		Expect(stackManager.CreateOrUpdateCall.Receives.StackName).To(Equal("concourse"))
		Expect(stackManager.CreateOrUpdateCall.Receives.Template).To(Equal(templates.Template{
			AWSTemplateFormatVersion: "some-template-version",
			Description:              "some-description",
		}))

		Expect(stackManager.WaitForCompletionCall.Receives.Client).To(Equal(cloudFormationClient))
		Expect(stackManager.WaitForCompletionCall.Receives.StackName).To(Equal("concourse"))
		Expect(stackManager.WaitForCompletionCall.Receives.SleepInterval).To(Equal(15 * time.Second))
	})

	Context("failure cases", func() {
		It("returns an error when stack can't be created or updated", func() {
			stackManager.CreateOrUpdateCall.Returns.Error = errors.New("stack create or update failed")

			_, err := creator.Create("some-key-pair-name", 0, cloudFormationClient)
			Expect(err).To(MatchError("stack create or update failed"))
		})

		It("returns an error when waiting for stack completion fails", func() {
			stackManager.WaitForCompletionCall.Returns.Error = errors.New("stack wait for completion failed")

			_, err := creator.Create("some-key-pair-name", 0, cloudFormationClient)
			Expect(err).To(MatchError("stack wait for completion failed"))
		})

		It("returns an error when describing the stack fails", func() {
			stackManager.DescribeCall.Returns.Error = errors.New("stack describe failed")

			_, err := creator.Create("some-key-pair-name", 0, cloudFormationClient)
			Expect(err).To(MatchError("stack describe failed"))
		})
	})
})
