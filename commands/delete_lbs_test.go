package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

var _ = Describe("Delete LBs", func() {
	var (
		deleteLBs                 commands.DeleteLBs
		awsCredentialValidator    *fakes.AWSCredentialValidator
		availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
		certificateManager        *fakes.CertificateManager
		infrastructureManager     *fakes.InfrastructureManager
		logger                    *fakes.Logger
		cloudConfigurator         *fakes.BoshCloudConfigurator
		cloudConfigManager        *fakes.CloudConfigManager
		boshClient                *fakes.BOSHClient
		boshClientProvider        *fakes.BOSHClientProvider
	)

	BeforeEach(func() {
		awsCredentialValidator = &fakes.AWSCredentialValidator{}
		availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}
		certificateManager = &fakes.CertificateManager{}
		infrastructureManager = &fakes.InfrastructureManager{}
		cloudConfigurator = &fakes.BoshCloudConfigurator{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		boshClient = &fakes.BOSHClient{}
		boshClientProvider = &fakes.BOSHClientProvider{}

		boshClientProvider.ClientCall.Returns.Client = boshClient

		logger = &fakes.Logger{}

		deleteLBs = commands.NewDeleteLBs(awsCredentialValidator, availabilityZoneRetriever,
			certificateManager, infrastructureManager, logger, cloudConfigurator, cloudConfigManager, boshClientProvider)
	})

	Describe("Execute", func() {
		It("updates cloud config", func() {
			availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"some-az"}
			infrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
				Name: "some-stack-name",
			}
			cloudConfigurator.ConfigureCall.Returns.CloudConfigInput = bosh.CloudConfigInput{
				AZs: []string{"some-az"},
				LBs: []bosh.LoadBalancerExtension{
					{
						Name: "some-lb",
					},
				},
			}

			_, err := deleteLBs.Execute([]string{}, storage.State{
				BOSH: storage.BOSH{
					DirectorAddress:  "some-director-address",
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
				},
				Stack: storage.Stack{
					Name: "some-stack-name",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))

			Expect(infrastructureManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))
			Expect(cloudConfigurator.ConfigureCall.Receives.Stack).To(Equal(cloudformation.Stack{
				Name: "some-stack-name",
			}))

			Expect(cloudConfigManager.UpdateCall.Receives.CloudConfigInput).To(Equal(bosh.CloudConfigInput{
				AZs: []string{"some-az"},
			}))
			Expect(cloudConfigManager.UpdateCall.Receives.BOSHClient).To(Equal(boshClient))
		})

		It("delete lbs from cloudformation and deletes certificate", func() {
			availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"a", "b", "c"}
			_, err := deleteLBs.Execute([]string{}, storage.State{
				Stack: storage.Stack{
					Name:            "some-stack",
					LBType:          "cf",
					CertificateName: "some-certificate",
				},
				AWS: storage.AWS{
					Region: "some-region",
				},
				KeyPair: storage.KeyPair{
					Name: "some-keypair",
				},
			})

			Expect(err).NotTo(HaveOccurred())

			Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(1))

			Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("some-region"))

			Expect(infrastructureManager.UpdateCall.Receives.KeyPairName).To(Equal("some-keypair"))
			Expect(infrastructureManager.UpdateCall.Receives.NumberOfAvailabilityZones).To(Equal(3))
			Expect(infrastructureManager.UpdateCall.Receives.StackName).To(Equal("some-stack"))
			Expect(infrastructureManager.UpdateCall.Receives.LBType).To(Equal(""))
			Expect(infrastructureManager.UpdateCall.Receives.LBCertificateARN).To(Equal(""))

			Expect(certificateManager.DeleteCall.Receives.CertificateName).To(Equal("some-certificate"))
		})

		Context("state management", func() {
			It("returns a state with no lb type nor certificate", func() {
				state, err := deleteLBs.Execute([]string{}, storage.State{
					Stack: storage.Stack{
						Name:            "some-stack",
						LBType:          "cf",
						CertificateName: "some-certificate",
					},
				})

				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{
					Stack: storage.Stack{
						Name: "some-stack",
					},
				}))
			})
		})

		Context("failure cases", func() {
			It("returns an error when aws credential validator fails to validate", func() {
				awsCredentialValidator.ValidateCall.Returns.Error = errors.New("validate failed")
				_, err := deleteLBs.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("validate failed"))
			})

			It("return an error when availability zone retriever fails to retrieve", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("retrieve failed")
				_, err := deleteLBs.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("retrieve failed"))
			})

			It("return an error when infrastructure manager fails to describe", func() {
				infrastructureManager.DescribeCall.Returns.Error = errors.New("describe failed")
				_, err := deleteLBs.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("describe failed"))
			})

			It("return an error when cloud config manager fails to update", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("update failed")
				_, err := deleteLBs.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("update failed"))
			})

			It("return an error when infrastructure manager fails to update", func() {
				infrastructureManager.UpdateCall.Returns.Error = errors.New("update failed")
				_, err := deleteLBs.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("update failed"))
			})

			It("return an error when certificate manager fails to delete", func() {
				certificateManager.DeleteCall.Returns.Error = errors.New("delete failed")
				_, err := deleteLBs.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("delete failed"))
			})
		})
	})
})
