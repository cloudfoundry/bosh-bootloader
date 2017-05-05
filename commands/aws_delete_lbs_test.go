package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
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
		environmentValidator      *fakes.EnvironmentValidator
		logger                    *fakes.Logger
		cloudConfigManager        *fakes.CloudConfigManager
		terraformManager          *fakes.TerraformManager
		stateStore                *fakes.StateStore

		incomingCloudformationState storage.State
		incomingTerraformState      storage.State
	)

	BeforeEach(func() {
		credentialValidator = &fakes.CredentialValidator{}
		availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}
		certificateManager = &fakes.CertificateManager{}
		infrastructureManager = &fakes.InfrastructureManager{}
		environmentValidator = &fakes.EnvironmentValidator{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		terraformManager = &fakes.TerraformManager{}
		stateStore = &fakes.StateStore{}

		logger = &fakes.Logger{}

		incomingCloudformationState = storage.State{
			Stack: storage.Stack{
				LBType:          "concourse",
				CertificateName: "some-certificate",
				Name:            "some-stack-name",
				BOSHAZ:          "some-bosh-az",
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

		incomingTerraformState = storage.State{
			TFState: "some-tf-state",
			LB: storage.LB{
				Type: "concourse",
				Cert: "some-cert",
				Key:  "some-key",
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
			certificateManager, infrastructureManager, logger, cloudConfigManager,
			stateStore, environmentValidator, terraformManager)
	})

	Describe("Execute", func() {
		Context("when the bbl env has a bosh director", func() {
			Context("when cloudformation is used for infrastructure", func() {
				It("updates cloud config", func() {
					availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"some-az"}
					infrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
						Name: "some-stack-name",
					}

					err := command.Execute(incomingCloudformationState)
					Expect(err).NotTo(HaveOccurred())

					Expect(infrastructureManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))

					Expect(cloudConfigManager.UpdateCall.Receives.State.Stack.LBType).To(Equal("none"))
				})

				It("delete lbs from cloudformation and deletes certificate", func() {
					availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"a", "b", "c"}

					err := command.Execute(incomingCloudformationState)
					Expect(err).NotTo(HaveOccurred())

					Expect(credentialValidator.ValidateCall.CallCount).To(Equal(1))

					Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("some-region"))

					Expect(infrastructureManager.UpdateCall.Receives.KeyPairName).To(Equal("some-keypair"))
					Expect(infrastructureManager.UpdateCall.Receives.AZs).To(Equal([]string{"a", "b", "c"}))
					Expect(infrastructureManager.UpdateCall.Receives.StackName).To(Equal("some-stack-name"))
					Expect(infrastructureManager.UpdateCall.Receives.LBType).To(Equal(""))
					Expect(infrastructureManager.UpdateCall.Receives.LBCertificateARN).To(Equal(""))
					Expect(infrastructureManager.UpdateCall.Receives.EnvID).To(Equal("some-env-id"))
					Expect(infrastructureManager.UpdateCall.Receives.BOSHAZ).To(Equal("some-bosh-az"))

					Expect(certificateManager.DeleteCall.Receives.CertificateName).To(Equal("some-certificate"))

					Expect(logger.StepCall.Messages).To(ContainElement("deleting certificate"))
				})

				It("returns an error if the environment validator fails", func() {
					environmentValidator.ValidateCall.Returns.Error = errors.New("failed to validate")
					err := command.Execute(incomingCloudformationState)
					Expect(err).To(MatchError("failed to validate"))
					Expect(environmentValidator.ValidateCall.Receives.State).To(Equal(incomingCloudformationState))
					Expect(environmentValidator.ValidateCall.CallCount).To(Equal(1))
				})
			})

			Context("when terraform is used for infrastructure", func() {
				It("updates cloud config", func() {
					err := command.Execute(incomingTerraformState)
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudConfigManager.UpdateCall.Receives.State.LB.Type).To(BeEmpty())
					Expect(cloudConfigManager.UpdateCall.Receives.State.LB.Cert).To(BeEmpty())
					Expect(cloudConfigManager.UpdateCall.Receives.State.LB.Key).To(BeEmpty())
				})

				It("runs terraform apply to delete lbs and certificate", func() {
					err := command.Execute(incomingTerraformState)
					Expect(err).NotTo(HaveOccurred())

					Expect(credentialValidator.ValidateCall.CallCount).To(Equal(1))

					Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal(""))

					Expect(infrastructureManager.UpdateCall.CallCount).To(Equal(0))

					Expect(certificateManager.DeleteCall.CallCount).To(Equal(0))

					Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))

					expectedTerraformState := incomingTerraformState
					expectedTerraformState.LB = storage.LB{}
					Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(expectedTerraformState))

					Expect(logger.StepCall.Messages).NotTo(ContainElement("deleting certificate"))
				})
			})
		})

		Context("when the bbl env was created without a bosh director", func() {
			It("does not try to update the cloud config", func() {
				state := storage.State{
					Stack: storage.Stack{
						LBType:          "concourse",
						CertificateName: "some-certificate",
						Name:            "some-stack-name",
						BOSHAZ:          "some-bosh-az",
					},
					NoDirector: true,
					AWS: storage.AWS{
						Region: "some-region",
					},
					KeyPair: storage.KeyPair{
						Name: "some-keypair",
					},
					EnvID: "some-env-id",
				}
				err := command.Execute(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})

		Context("when cloudformation is used for infrastructure", func() {
			It("returns an error if there is no lb", func() {
				err := command.Execute(storage.State{
					Stack: storage.Stack{
						LBType: "none",
					},
				})
				Expect(err).To(MatchError(commands.LBNotFound))
			})
		})

		Context("when terraform is used for infrastructure", func() {
			It("returns an error if there is no lb", func() {
				err := command.Execute(storage.State{
					TFState: "some-tf-state",
				})
				Expect(err).To(MatchError(commands.LBNotFound))
			})
		})

		Context("state management", func() {
			It("saves state with no lb type before deleting certificate", func() {
				certificateManager.DeleteCall.Returns.Error = errors.New("failed to delete")
				err := command.Execute(storage.State{
					Stack: storage.Stack{
						Name:            "some-stack",
						LBType:          "cf",
						CertificateName: "some-certificate",
					},
				})
				Expect(err).To(MatchError("failed to delete"))

				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
					Stack: storage.Stack{
						Name:            "some-stack",
						LBType:          "none",
						CertificateName: "some-certificate",
					},
				}))
			})

			It("saves state with no lb type nor certificate", func() {
				err := command.Execute(storage.State{
					Stack: storage.Stack{
						Name:            "some-stack",
						LBType:          "cf",
						CertificateName: "some-certificate",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.SetCall.CallCount).To(Equal(2))
				Expect(stateStore.SetCall.Receives[1].State).To(Equal(storage.State{
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
				credentialValidator.ValidateCall.Returns.Error = errors.New("validate failed")
				err := command.Execute(incomingCloudformationState)
				Expect(err).To(MatchError("validate failed"))
			})

			It("return an error when availability zone retriever fails to retrieve", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("retrieve failed")
				err := command.Execute(incomingCloudformationState)
				Expect(err).To(MatchError("retrieve failed"))
			})

			Context("when terraform manager fails to apply with terraformManagerError", func() {
				It("return an error", func() {
					terraformManager.ApplyCall.Returns.Error = errors.New("apply failed")
					err := command.Execute(incomingTerraformState)
					Expect(err).To(MatchError("apply failed"))
				})
			})

			Context("when terraform manager fails to apply with non-terraformManagerError", func() {
				var (
					managerError *fakes.TerraformManagerError
				)

				BeforeEach(func() {
					managerError = &fakes.TerraformManagerError{}
					managerError.BBLStateCall.Returns.BBLState = storage.State{
						TFState: "some-partial-tf-state",
					}
					managerError.ErrorCall.Returns = "cannot apply"
					terraformManager.ApplyCall.Returns.Error = managerError
				})

				It("return an error", func() {
					err := command.Execute(incomingTerraformState)
					Expect(err).To(MatchError("cannot apply"))

					Expect(stateStore.SetCall.CallCount).To(Equal(1))
					Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
						TFState: "some-partial-tf-state",
					}))
				})

				Context("when the terraform manager error fails to return a bbl state", func() {
					BeforeEach(func() {
						managerError.BBLStateCall.Returns.Error = errors.New("failed to retrieve bbl state")
					})

					It("saves the bbl state and returns the error", func() {
						err := command.Execute(incomingTerraformState)
						Expect(err).To(MatchError("the following errors occurred:\ncannot apply,\nfailed to retrieve bbl state"))
					})
				})
			})

			It("return an error when infrastructure manager fails to describe", func() {
				infrastructureManager.DescribeCall.Returns.Error = errors.New("describe failed")
				err := command.Execute(incomingCloudformationState)
				Expect(err).To(MatchError("describe failed"))
			})

			It("return an error when cloud config manager fails to update", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("update failed")
				err := command.Execute(incomingCloudformationState)
				Expect(err).To(MatchError("update failed"))
			})

			It("return an error when infrastructure manager fails to update", func() {
				infrastructureManager.UpdateCall.Returns.Error = errors.New("update failed")
				err := command.Execute(incomingCloudformationState)
				Expect(err).To(MatchError("update failed"))
			})

			It("return an error when certificate manager fails to delete", func() {
				certificateManager.DeleteCall.Returns.Error = errors.New("delete failed")
				err := command.Execute(incomingCloudformationState)
				Expect(err).To(MatchError("delete failed"))
			})

			It("returns an error when the state fails to save lb type", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to save state")}}
				err := command.Execute(incomingCloudformationState)
				Expect(err).To(MatchError("failed to save state"))
			})
			It("returns an error when the state fails to save certificate deletion", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to save state")}}
				err := command.Execute(incomingCloudformationState)
				Expect(err).To(MatchError("failed to save state"))
			})
		})
	})
})
