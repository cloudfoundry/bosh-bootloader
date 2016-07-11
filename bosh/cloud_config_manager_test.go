package bosh_test

import (
	"errors"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudConfigManager", func() {
	Describe("Update", func() {
		var (
			logger               *fakes.Logger
			cloudConfigGenerator *fakes.CloudConfigGenerator
			boshClient           *fakes.BOSHClient
			cloudConfigManager   bosh.CloudConfigManager
			cloudConfigInput     bosh.CloudConfigInput
		)

		BeforeEach(func() {
			logger = &fakes.Logger{}
			cloudConfigGenerator = &fakes.CloudConfigGenerator{}
			boshClient = &fakes.BOSHClient{}
			cloudConfigManager = bosh.NewCloudConfigManager(logger, cloudConfigGenerator)

			cloudConfigGenerator.GenerateCall.Returns.CloudConfig = bosh.CloudConfig{
				VMTypes: []bosh.VMType{
					{
						Name: "some-vm-type",
					},
				},
			}

			cloudConfigInput = bosh.CloudConfigInput{
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
						SecurityGroups: []string{"some-internal-security-group"},
					},
					{
						AZ:             "us-east-1b",
						Subnet:         "some-internal-subnet-2",
						CIDR:           "some-cidr-block-2",
						SecurityGroups: []string{"some-internal-security-group"},
					},
					{
						AZ:             "us-east-1c",
						Subnet:         "some-internal-subnet-3",
						CIDR:           "some-cidr-block-3",
						SecurityGroups: []string{"some-internal-security-group"},
					},
					{
						AZ:             "us-east-1e",
						Subnet:         "some-internal-subnet-4",
						CIDR:           "some-cidr-block-4",
						SecurityGroups: []string{"some-internal-security-group"},
					},
				},
				LBs: []bosh.LoadBalancerExtension{},
			}
		})

		It("generates and applies a cloud config", func() {
			err := cloudConfigManager.Update(cloudConfigInput, boshClient)
			Expect(err).NotTo(HaveOccurred())

			//Expect(logger.StepCall.Receives.Message).To(Equal("generating cloud config"))
			manifestYAML, err := yaml.Marshal(bosh.CloudConfig{
				VMTypes: []bosh.VMType{
					{
						Name: "some-vm-type",
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.UpdateCloudConfigCall.CallCount).To(Equal(1))
			Expect(boshClient.UpdateCloudConfigCall.Receives.Yaml).To(Equal(manifestYAML))

			Expect(cloudConfigGenerator.GenerateCall.Receives.CloudConfigInput).To(Equal(cloudConfigInput))
			//Expect(logger.StepCall.Receives.Message).To(Equal("applying cloud config"))
		})

		Context("failure cases", func() {
			It("returns an error when the generate fails", func() {
				cloudConfigGenerator.GenerateCall.Returns.Error = errors.New("generate failed")
				err := cloudConfigManager.Update(cloudConfigInput, boshClient)

				Expect(err).To(MatchError("generate failed"))
			})

			It("returns an error when the bosh client fails to upload the cloud config", func() {
				boshClient.UpdateCloudConfigCall.Returns.Error = errors.New("failed to upload")
				err := cloudConfigManager.Update(cloudConfigInput, boshClient)

				Expect(err).To(MatchError("failed to upload"))
			})
		})
	})
})
