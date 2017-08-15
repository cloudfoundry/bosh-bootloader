package cloudformation_test

import (
	"errors"
	"time"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InfrastructureManager", func() {
	var (
		builder               *fakes.TemplateBuilder
		stackManager          *fakes.StackManager
		infrastructureManager cloudformation.InfrastructureManager

		azs []string
	)

	BeforeEach(func() {

		builder = &fakes.TemplateBuilder{}
		builder.BuildCall.Returns.Template = templates.Template{
			AWSTemplateFormatVersion: "some-template-version",
			Description:              "some-description",
		}

		stackManager = &fakes.StackManager{}

		infrastructureManager = cloudformation.NewInfrastructureManager(builder, stackManager)

		azs = []string{"some-zone-1", "some-zone-2"}
	})

	Describe("Update", func() {
		BeforeEach(func() {
			stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{Name: "some-stack-name"}
		})

		It("updates the stack and returns the stack", func() {
			stackManager.GetPhysicalIDForResourceCall.Returns.PhysicalResourceID = "some-bosh-user-id"

			stack, err := infrastructureManager.Update("some-key-pair-name", azs, "some-stack-name", "some-bosh-az", "some-lb-type", "some-lb-certificate-arn", "some-env-id-time:stamp")
			Expect(err).NotTo(HaveOccurred())

			Expect(stackManager.GetPhysicalIDForResourceCall.Receives.StackName).To(Equal("some-stack-name"))
			Expect(stackManager.GetPhysicalIDForResourceCall.Receives.LogicalResourceID).To(Equal("BOSHUser"))

			Expect(stack).To(Equal(cloudformation.Stack{Name: "some-stack-name"}))
			Expect(builder.BuildCall.Receives.KeyPairName).To(Equal("some-key-pair-name"))
			Expect(builder.BuildCall.Receives.AZs).To(Equal(azs))
			Expect(builder.BuildCall.Receives.LBType).To(Equal("some-lb-type"))
			Expect(builder.BuildCall.Receives.LBCertificateARN).To(Equal("some-lb-certificate-arn"))
			Expect(builder.BuildCall.Receives.IAMUserName).To(Equal("some-bosh-user-id"))
			Expect(builder.BuildCall.Receives.EnvID).To(Equal("some-env-id-time:stamp"))
			Expect(builder.BuildCall.Receives.BOSHAZ).To(Equal("some-bosh-az"))

			Expect(stackManager.UpdateCall.Receives.StackName).To(Equal("some-stack-name"))
			Expect(stackManager.UpdateCall.Receives.Template).To(Equal(templates.Template{
				AWSTemplateFormatVersion: "some-template-version",
				Description:              "some-description",
			}))
			Expect(stackManager.UpdateCall.Receives.Tags).To(Equal(cloudformation.Tags{
				{
					Key:   "bbl-env-id",
					Value: "some-env-id-time:stamp",
				},
			}))

			Expect(stackManager.WaitForCompletionCall.Receives.StackName).To(Equal("some-stack-name"))
			Expect(stackManager.WaitForCompletionCall.Receives.SleepInterval).To(Equal(15 * time.Second))
			Expect(stackManager.WaitForCompletionCall.Receives.Action).To(Equal("applying cloudformation template"))

			Expect(stackManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))
		})

		Context("failure cases", func() {
			It("returns an error when it cannot get physical id for BOSHUser", func() {
				stackManager.GetPhysicalIDForResourceCall.Returns.Error = errors.New("failed to get physical id for resource")

				_, err := infrastructureManager.Update("some-key-pair-name", azs, "some-stack-name", "some-bosh-az", "some-lb-type", "some-lb-certificate-arn", "some-env-id-time:stamp")
				Expect(err).To(MatchError("failed to get physical id for resource"))
			})

			It("returns an error when the update stack call fails", func() {
				stackManager.UpdateCall.Returns.Error = errors.New("stack update call failed")

				_, err := infrastructureManager.Update("some-key-pair-name", azs, "some-stack-name", "some-bosh-az", "some-lb-type", "some-lb-certificate-arn", "some-env-id-time:stamp")
				Expect(err).To(MatchError("stack update call failed"))
			})

			It("returns an error when the wait for completion call fails", func() {
				stackManager.WaitForCompletionCall.Returns.Error = errors.New("failed to wait for completion")

				_, err := infrastructureManager.Update("some-key-pair-name", azs, "some-stack-name", "some-bosh-az", "some-lb-type", "some-lb-certificate-arn", "some-env-id-time:stamp")
				Expect(err).To(MatchError("failed to wait for completion"))
			})
		})
	})

	Describe("Exists", func() {
		It("returns true when the stack exists", func() {
			stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{}

			exists, err := infrastructureManager.Exists("some-stack-name")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("returns false when the stack does not exist", func() {
			stackManager.DescribeCall.Returns.Error = cloudformation.StackNotFound

			exists, err := infrastructureManager.Exists("some-stack-name")

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		Describe("failure cases", func() {
			It("returns an error when the stack manager returns a different error", func() {
				stackManager.DescribeCall.Returns.Error = errors.New("some other error")
				_, err := infrastructureManager.Exists("some-stack-name")
				Expect(err).To(MatchError("some other error"))
			})
		})
	})

	Describe("Delete", func() {
		It("deletes the underlying infrastructure", func() {
			err := infrastructureManager.Delete("some-stack-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(stackManager.DeleteCall.Receives.StackName).To(Equal("some-stack-name"))

			Expect(stackManager.WaitForCompletionCall.Receives.StackName).To(Equal("some-stack-name"))
			Expect(stackManager.WaitForCompletionCall.Receives.SleepInterval).To(Equal(15 * time.Second))
			Expect(stackManager.WaitForCompletionCall.Receives.Action).To(Equal("deleting cloudformation stack"))
		})

		Context("when the stack goes away after being deleted", func() {
			It("returns without an error", func() {
				stackManager.WaitForCompletionCall.Returns.Error = cloudformation.StackNotFound

				err := infrastructureManager.Delete("some-stack-name")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("failure cases", func() {
			Context("when the stack fails to delete", func() {
				It("returns an error", func() {
					stackManager.DeleteCall.Returns.Error = errors.New("failed to delete stack")

					err := infrastructureManager.Delete("some-stack-name")
					Expect(err).To(MatchError("failed to delete stack"))
				})
			})
			Context("when the waiting for completion fails", func() {
				It("returns an error", func() {
					stackManager.WaitForCompletionCall.Returns.Error = errors.New("wait for completion failed")

					err := infrastructureManager.Delete("some-stack-name")
					Expect(err).To(MatchError("wait for completion failed"))
				})
			})
		})
	})

	Describe("Describe", func() {
		It("returns a stack with a given name", func() {
			expectedStack := cloudformation.Stack{
				Name:   "some-stack-name",
				Status: "some-status",
				Outputs: map[string]string{
					"some-output":       "some-value",
					"some-other-output": "some-other-value",
				},
			}

			stackManager.DescribeCall.Returns.Stack = expectedStack

			stack, err := infrastructureManager.Describe("some-stack-name")
			Expect(err).NotTo(HaveOccurred())
			Expect(stack).To(Equal(expectedStack))

			Expect(stackManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))
		})
	})
})
