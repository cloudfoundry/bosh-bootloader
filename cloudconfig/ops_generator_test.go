package cloudconfig_test

import (
	"errors"
	"fmt"

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
			awsOpsGenerator *fakes.CloudConfigOpsGenerator
			gcpOpsGenerator *fakes.CloudConfigOpsGenerator
			opsGenerator    cloudconfig.OpsGenerator

			incomingState storage.State
		)

		BeforeEach(func() {
			awsOpsGenerator = &fakes.CloudConfigOpsGenerator{}
			gcpOpsGenerator = &fakes.CloudConfigOpsGenerator{}

			awsOpsGenerator.GenerateCall.Returns.OpsYAML = "some-aws-ops"
			gcpOpsGenerator.GenerateCall.Returns.OpsYAML = "some-gcp-ops"
			opsGenerator = cloudconfig.NewOpsGenerator(awsOpsGenerator, gcpOpsGenerator)
		})

		DescribeTable("returns an ops file to transform base cloud config to iaas specific cloud config", func(iaas string) {
			incomingState = storage.State{
				IAAS: iaas,
			}
			expectedIAAS := fmt.Sprintf("some-%s-ops", iaas)

			opsYAML, err := opsGenerator.Generate(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(opsYAML).To(Equal(expectedIAAS))
		},
			Entry("when iaas is gcp", "gcp"),
			Entry("when iaas is aws", "aws"),
		)

		Context("failure cases", func() {
			It("returns an error if iaas is invalid", func() {
				incomingState = storage.State{
					IAAS: "invalid-iaas",
				}
				_, err := opsGenerator.Generate(incomingState)
				Expect(err).To(MatchError("invalid iaas type"))
			})

			DescribeTable("returns an error when it fails to generate iaas cloud config", func(iaas string) {
				awsOpsGenerator.GenerateCall.Returns.Error = errors.New("failed to generate aws cloud config")
				gcpOpsGenerator.GenerateCall.Returns.Error = errors.New("failed to generate gcp cloud config")
				incomingState = storage.State{
					IAAS: iaas,
				}
				expectedError := fmt.Errorf("failed to generate %s cloud config", iaas)

				_, err := opsGenerator.Generate(incomingState)
				Expect(err).To(MatchError(expectedError))
			},
				Entry("when iaas is gcp", "gcp"),
				Entry("when iaas is aws", "aws"),
			)
		})
	})
})
