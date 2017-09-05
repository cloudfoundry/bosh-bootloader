package cloudconfig_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("OpsGenerator", func() {
	Describe("Generate", func() {
		var (
			awsCloudFormationOpsGenerator *fakes.CloudConfigOpsGenerator
			awsTerraformOpsGenerator      *fakes.CloudConfigOpsGenerator
			gcpOpsGenerator               *fakes.CloudConfigOpsGenerator
			azureOpsGenerator             *fakes.CloudConfigOpsGenerator
			opsGenerator                  cloudconfig.OpsGenerator

			incomingState storage.State
		)

		BeforeEach(func() {
			awsCloudFormationOpsGenerator = &fakes.CloudConfigOpsGenerator{}
			awsTerraformOpsGenerator = &fakes.CloudConfigOpsGenerator{}
			gcpOpsGenerator = &fakes.CloudConfigOpsGenerator{}
			azureOpsGenerator = &fakes.CloudConfigOpsGenerator{}

			awsCloudFormationOpsGenerator.GenerateCall.Returns.OpsYAML = "some-aws-cloudformation-ops"
			awsTerraformOpsGenerator.GenerateCall.Returns.OpsYAML = "some-aws-terraform-ops"
			gcpOpsGenerator.GenerateCall.Returns.OpsYAML = "some-gcp-ops"
			azureOpsGenerator.GenerateCall.Returns.OpsYAML = "some-azure-ops"

			opsGenerator = cloudconfig.NewOpsGenerator(awsCloudFormationOpsGenerator, awsTerraformOpsGenerator, gcpOpsGenerator, azureOpsGenerator)
		})

		DescribeTable("returns an ops file to transform base cloud config to iaas specific cloud config", func(incomingState storage.State, expectedOpsYAML string) {
			opsYAML, err := opsGenerator.Generate(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(opsYAML).To(Equal(expectedOpsYAML))
		},
			Entry("when iaas is gcp", storage.State{
				IAAS: "gcp",
			}, "some-gcp-ops"),
			Entry("when iaas is aws and terraform was used to create infrastructure", storage.State{
				IAAS:    "aws",
				TFState: "some-tf-state",
			}, "some-aws-terraform-ops"),
			Entry("when iaas is aws and cloudformation was used to create infrastructure", storage.State{
				IAAS:    "aws",
				TFState: "",
			}, "some-aws-cloudformation-ops"),
			Entry("when iaas is azure", storage.State{
				IAAS: "azure",
			}, "some-azure-ops"),
		)

		Context("failure cases", func() {
			It("returns an error if iaas is invalid", func() {
				incomingState = storage.State{
					IAAS: "invalid-iaas",
				}
				_, err := opsGenerator.Generate(incomingState)
				Expect(err).To(MatchError("invalid iaas type"))
			})

			DescribeTable("returns an error when it fails to generate iaas cloud config", func(incomingState storage.State, getOpsGenerator func() *fakes.CloudConfigOpsGenerator) {
				getOpsGenerator().GenerateCall.Returns.Error = errors.New("failed to generate cloud config")

				_, err := opsGenerator.Generate(incomingState)
				Expect(err).To(MatchError("failed to generate cloud config"))
			},
				Entry("when iaas is gcp", storage.State{
					IAAS: "gcp",
				}, func() *fakes.CloudConfigOpsGenerator {
					return gcpOpsGenerator
				}),
				Entry("when iaas is aws and terraform was used to create infrastructure", storage.State{
					IAAS:    "aws",
					TFState: "some-tf-state",
				}, func() *fakes.CloudConfigOpsGenerator {
					return awsTerraformOpsGenerator
				}),
				Entry("when iaas is aws and cloudformation was used to create infrastructure", storage.State{
					IAAS:    "aws",
					TFState: "",
				}, func() *fakes.CloudConfigOpsGenerator {
					return awsCloudFormationOpsGenerator
				}),
				Entry("when iaas is azure", storage.State{
					IAAS: "azure",
				}, func() *fakes.CloudConfigOpsGenerator {
					return azureOpsGenerator
				}),
			)
		})
	})
})
