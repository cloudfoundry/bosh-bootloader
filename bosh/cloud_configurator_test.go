package bosh_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudConfigurator", func() {
	Describe("Configure", func() {
		var (
			logger               *fakes.Logger
			cloudConfigGenerator *fakes.CloudConfigGenerator
			boshClient           *fakes.BOSHClient
			cloudFormationStack  cloudformation.Stack
			azs                  []string
			cloudConfigurator    bosh.CloudConfigurator
		)

		BeforeEach(func() {
			logger = &fakes.Logger{}
			cloudConfigGenerator = &fakes.CloudConfigGenerator{}
			boshClient = &fakes.BOSHClient{}
			cloudConfigurator = bosh.NewCloudConfigurator(logger, cloudConfigGenerator)

			cloudConfigGenerator.GenerateCall.Returns.CloudConfig = bosh.CloudConfig{
				VMTypes: []bosh.VMType{
					{
						Name: "some-vm-type",
					},
				},
			}

			cloudFormationStack = cloudformation.Stack{
				Outputs: map[string]string{
					"InternalSubnet1AZ":            "us-east-1a",
					"InternalSubnet2AZ":            "us-east-1b",
					"InternalSubnet3AZ":            "us-east-1c",
					"InternalSubnet4AZ":            "us-east-1e",
					"InternalSubnet1Name":          "some-internal-subnet-1",
					"InternalSubnet2Name":          "some-internal-subnet-2",
					"InternalSubnet3Name":          "some-internal-subnet-3",
					"InternalSubnet4Name":          "some-internal-subnet-4",
					"InternalSubnet1CIDR":          "some-cidr-block-1",
					"InternalSubnet2CIDR":          "some-cidr-block-2",
					"InternalSubnet3CIDR":          "some-cidr-block-3",
					"InternalSubnet4CIDR":          "some-cidr-block-4",
					"InternalSubnet1SecurityGroup": "some-security-group-1",
					"InternalSubnet2SecurityGroup": "some-security-group-2",
					"InternalSubnet3SecurityGroup": "some-security-group-3",
					"InternalSubnet4SecurityGroup": "some-security-group-4",
					"BOSHEIP":                      "bosh-director-url",
				},
			}

			azs = []string{"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1e"}
		})

		It("generates a bosh cloud config", func() {
			err := cloudConfigurator.Configure(cloudFormationStack, azs, boshClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Receives.Message).To(Equal("applying cloud config"))
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
					{
						AZ:             "us-east-1e",
						Subnet:         "some-internal-subnet-4",
						CIDR:           "some-cidr-block-4",
						SecurityGroups: []string{"some-security-group-4"},
					},
				},
				LBs: map[string]string{},
			}))
		})

		It("applies the generated cloud config", func() {
			err := cloudConfigurator.Configure(cloudFormationStack, azs, boshClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Receives.Message).To(Equal("applying cloud config"))

			yaml, err := candiedyaml.Marshal(bosh.CloudConfig{
				VMTypes: []bosh.VMType{
					{
						Name: "some-vm-type",
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.UpdateCloudConfigCall.CallCount).To(Equal(1))
			Expect(boshClient.UpdateCloudConfigCall.Receives.Yaml).To(Equal(yaml))
		})

		Context("vm extensions", func() {
			Context("when there is no lb", func() {
				It("generates a cloud config with no lb vm extension", func() {
					cloudFormationStack.Outputs["ConcourseLoadBalancer"] = ""
					err := cloudConfigurator.Configure(cloudFormationStack, azs, boshClient)
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudConfigGenerator.GenerateCall.Receives.CloudConfigInput.LBs).To(HaveLen(0))
				})
			})

			Context("when the load balancer type is concourse", func() {
				It("generates a cloud config with a concourse lb vm extension", func() {
					cloudFormationStack.Outputs["ConcourseLoadBalancer"] = "some-lb"

					err := cloudConfigurator.Configure(cloudFormationStack, azs, boshClient)
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudConfigGenerator.GenerateCall.Receives.CloudConfigInput.LBs).To(HaveKeyWithValue("lb", "some-lb"))
				})
			})

			Context("when the load balancer type is cf", func() {
				It("generates a cloud config with a cf lb vm extension", func() {
					cloudFormationStack.Outputs["CFRouterLoadBalancer"] = "some-cf-lb"

					err := cloudConfigurator.Configure(cloudFormationStack, azs, boshClient)
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudConfigGenerator.GenerateCall.Receives.CloudConfigInput.LBs).To(HaveKeyWithValue("cf-lb", "some-cf-lb"))
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when cloud config cannot be applied", func() {
				boshClient.UpdateCloudConfigCall.Returns.Error = errors.New("failed to apply")
				err := cloudConfigurator.Configure(cloudFormationStack, azs, boshClient)
				Expect(err).To(MatchError("failed to apply"))
			})

			It("returns an error when cloud config cannot be generated", func() {
				cloudConfigGenerator.GenerateCall.Returns.Error = errors.New("cloud config generator failed")
				err := cloudConfigurator.Configure(cloudformation.Stack{}, []string{}, boshClient)

				Expect(err).To(MatchError("cloud config generator failed"))
			})
		})
	})
})
