package bosh_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("CloudConfigurator", func() {
	Describe("Configure", func() {
		var (
			logger               *fakes.Logger
			cloudConfigGenerator *fakes.CloudConfigGenerator
			cloudConfigurator    bosh.CloudConfigurator
		)

		BeforeEach(func() {
			logger = &fakes.Logger{}
			cloudConfigGenerator = &fakes.CloudConfigGenerator{}
			cloudConfigurator = bosh.NewCloudConfigurator(logger, cloudConfigGenerator)
		})

		It("Generates and prints a bosh cloud config", func() {
			cloudConfigGenerator.GenerateCall.Returns.CloudConfig = bosh.CloudConfig{
				VMTypes: []bosh.VMType{
					{
						Name: "some-vm-type",
					},
				},
			}

			cloudformationStack := cloudformation.Stack{
				Outputs: map[string]string{
					"InternalSubnet1AZ":            "us-east-1a",
					"InternalSubnet2AZ":            "us-east-1b",
					"InternalSubnet3AZ":            "us-east-1c",
					"InternalSubnet1Name":          "some-internal-subnet-1",
					"InternalSubnet2Name":          "some-internal-subnet-2",
					"InternalSubnet3Name":          "some-internal-subnet-3",
					"InternalSubnet1CIDR":          "some-cidr-block-1",
					"InternalSubnet2CIDR":          "some-cidr-block-2",
					"InternalSubnet3CIDR":          "some-cidr-block-3",
					"InternalSubnet1SecurityGroup": "some-security-group-1",
					"InternalSubnet2SecurityGroup": "some-security-group-2",
					"InternalSubnet3SecurityGroup": "some-security-group-3",
				},
			}
			azs := []string{"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1e"}
			err := cloudConfigurator.Configure(cloudformationStack, azs)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.StepCall.Receives.Message).To(Equal("generating cloud config"))
			Expect(logger.PrintlnCall.Receives.Message).To(MatchYAML(`---
vm_types:
- name: some-vm-type
`))
			Expect(cloudConfigGenerator.GenerateCall.CallCount).To(Equal(1))

			Expect(cloudConfigGenerator.GenerateCall.Receives.CloudConfigInput).To(Equal(bosh.CloudConfigInput{
				AZs: []string{
					"us-east-1a",
					"us-east-1b",
					"us-east-1c",
					"us-east-1e",
				},
				Subnets: []bosh.SubnetInput{
					{
						AZ:             "us-east-1a",
						Subnet:         "some-internal-subnet-1",
						CIDR:           "some-cidr-block-1",
						SecurityGroups: []string{"some-security-group-1"},
					},
					{
						AZ:             "us-east-1b",
						Subnet:         "some-internal-subnet-2",
						CIDR:           "some-cidr-block-2",
						SecurityGroups: []string{"some-security-group-2"},
					},
					{
						AZ:             "us-east-1c",
						Subnet:         "some-internal-subnet-3",
						CIDR:           "some-cidr-block-3",
						SecurityGroups: []string{"some-security-group-3"},
					},
				},
			}))
		})

		Context("failure cases", func() {
			It("returns an error when cloud config cannot be generated", func() {
				cloudConfigGenerator.GenerateCall.Returns.Error = errors.New("cloud config generator failed")
				err := cloudConfigurator.Configure(cloudformation.Stack{}, []string{})

				Expect(err).To(MatchError("cloud config generator failed"))
			})
		})
	})
})
