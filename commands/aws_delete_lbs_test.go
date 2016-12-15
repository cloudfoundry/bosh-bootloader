package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete LBs", func() {
	var (
		command                   commands.AWSDeleteLBs
		credentialValidator       *fakes.CredentialValidator
		availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
		certificateManager        *fakes.CertificateManager
		infrastructureManager     *fakes.InfrastructureManager
		logger                    *fakes.Logger
		cloudConfigurator         *fakes.BoshCloudConfigurator
		cloudConfigManager        *fakes.CloudConfigManager
		boshClient                *fakes.BOSHClient
		boshClientProvider        *fakes.BOSHClientProvider
		stateStore                *fakes.StateStore
		incomingState             storage.State
	)

	BeforeEach(func() {
		credentialValidator = &fakes.CredentialValidator{}
		availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}
		certificateManager = &fakes.CertificateManager{}
		infrastructureManager = &fakes.InfrastructureManager{}
		cloudConfigurator = &fakes.BoshCloudConfigurator{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		boshClient = &fakes.BOSHClient{}
		boshClientProvider = &fakes.BOSHClientProvider{}
		stateStore = &fakes.StateStore{}

		boshClientProvider.ClientCall.Returns.Client = boshClient

		logger = &fakes.Logger{}

		incomingState = storage.State{
			Stack: storage.Stack{
				LBType:          "concourse",
				CertificateName: "some-certificate",
				Name:            "some-stack-name",
			},
			BOSH: storage.BOSH{
				DirectorAddress:  "some-director-address",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
			},
			AWS: storage.AWS{
				Region: "some-region",
			},
			KeyPair: storage.KeyPair{
				Name: "some-keypair",
			},
			EnvID: "some-env-id",
		}

		infrastructureManager.ExistsCall.Returns.Exists = true

		command = commands.NewAWSDeleteLBs(credentialValidator, availabilityZoneRetriever,
			certificateManager, infrastructureManager, logger, cloudConfigurator, cloudConfigManager,
			boshClientProvider, stateStore)
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

			err := command.Execute(incomingState)
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
			err := command.Execute(incomingState)

			Expect(err).NotTo(HaveOccurred())

			Expect(credentialValidator.ValidateAWSCall.CallCount).To(Equal(1))

			Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("some-region"))

			Expect(infrastructureManager.UpdateCall.Receives.KeyPairName).To(Equal("some-keypair"))
			Expect(infrastructureManager.UpdateCall.Receives.NumberOfAvailabilityZones).To(Equal(3))
			Expect(infrastructureManager.UpdateCall.Receives.StackName).To(Equal("some-stack-name"))
			Expect(infrastructureManager.UpdateCall.Receives.LBType).To(Equal(""))
			Expect(infrastructureManager.UpdateCall.Receives.LBCertificateARN).To(Equal(""))
			Expect(infrastructureManager.UpdateCall.Receives.EnvID).To(Equal("some-env-id"))

			Expect(certificateManager.DeleteCall.Receives.CertificateName).To(Equal("some-certificate"))

			Expect(logger.StepCall.Messages).To(ContainElement("deleting certificate"))
		})

		It("checks if the bosh director exists", func() {
			err := command.Execute(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))

			Expect(boshClient.InfoCall.CallCount).To(Equal(1))
		})

		Context("if the user hasn't bbl'd up yet", func() {
			It("returns an error if the stack does not exist", func() {
				infrastructureManager.ExistsCall.Returns.Exists = false
				err := command.Execute(storage.State{})
				Expect(err).To(MatchError(commands.BBLNotFound))
			})

			It("returns an error if the bosh director does not exist", func() {
				boshClient.InfoCall.Returns.Error = errors.New("director not found")

				err := command.Execute(incomingState)
				Expect(err).To(MatchError(commands.BBLNotFound))
			})
		})

		It("returns an error if there is no lb", func() {
			err := command.Execute(storage.State{
				Stack: storage.Stack{
					LBType: "none",
				},
			})
			Expect(err).To(MatchError(commands.LBNotFound))
		})

		Context("state management", func() {
			It("saves state with no lb type nor certificate", func() {
				err := command.Execute(storage.State{
					Stack: storage.Stack{
						Name:            "some-stack",
						LBType:          "cf",
						CertificateName: "some-certificate",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
					Stack: storage.Stack{
						Name:            "some-stack",
						LBType:          "none",
						CertificateName: "",
					},
				}))
			})
		})

		Context("failure cases", func() {
			It("returns an error when aws credential validator fails to validate", func() {
				credentialValidator.ValidateAWSCall.Returns.Error = errors.New("validate failed")
				err := command.Execute(incomingState)
				Expect(err).To(MatchError("validate failed"))
			})

			It("return an error when availability zone retriever fails to retrieve", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("retrieve failed")
				err := command.Execute(incomingState)
				Expect(err).To(MatchError("retrieve failed"))
			})

			It("return an error when infrastructure manager fails to describe", func() {
				infrastructureManager.DescribeCall.Returns.Error = errors.New("describe failed")
				err := command.Execute(incomingState)
				Expect(err).To(MatchError("describe failed"))
			})

			It("return an error when cloud config manager fails to update", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("update failed")
				err := command.Execute(incomingState)
				Expect(err).To(MatchError("update failed"))
			})

			It("return an error when infrastructure manager fails to update", func() {
				infrastructureManager.UpdateCall.Returns.Error = errors.New("update failed")
				err := command.Execute(incomingState)
				Expect(err).To(MatchError("update failed"))
			})

			It("return an error when certificate manager fails to delete", func() {
				certificateManager.DeleteCall.Returns.Error = errors.New("delete failed")
				err := command.Execute(incomingState)
				Expect(err).To(MatchError("delete failed"))
			})

			It("returns an error when the state fails to be saved", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to save state")}}
				err := command.Execute(incomingState)
				Expect(err).To(MatchError("failed to save state"))
			})
		})
	})
})
