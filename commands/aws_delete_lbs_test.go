package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete LBs", func() {
	var (
		command              commands.AWSDeleteLBs
		credentialValidator  *fakes.CredentialValidator
		environmentValidator *fakes.EnvironmentValidator
		logger               *fakes.Logger
		cloudConfigManager   *fakes.CloudConfigManager
		terraformManager     *fakes.TerraformManager
		stateStore           *fakes.StateStore

		incomingState storage.State
	)

	BeforeEach(func() {
		credentialValidator = &fakes.CredentialValidator{}
		environmentValidator = &fakes.EnvironmentValidator{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		terraformManager = &fakes.TerraformManager{}
		stateStore = &fakes.StateStore{}

		logger = &fakes.Logger{}

		incomingState = storage.State{
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

		command = commands.NewAWSDeleteLBs(credentialValidator,
			logger, cloudConfigManager,
			stateStore, environmentValidator, terraformManager)
	})

	Describe("Execute", func() {
		Context("when the bbl env has a bosh director", func() {
			It("updates cloud config", func() {
				err := command.Execute(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.UpdateCall.Receives.State.LB.Type).To(BeEmpty())
				Expect(cloudConfigManager.UpdateCall.Receives.State.LB.Cert).To(BeEmpty())
				Expect(cloudConfigManager.UpdateCall.Receives.State.LB.Key).To(BeEmpty())
			})

			It("runs terraform apply to delete lbs and certificate", func() {
				err := command.Execute(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(credentialValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))

				expectedTerraformState := incomingState
				expectedTerraformState.LB = storage.LB{}
				Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(expectedTerraformState))
			})
		})

		Context("when the bbl env was created without a bosh director", func() {
			It("does not try to update the cloud config", func() {
				state := storage.State{
					Stack: storage.Stack{
						CertificateName: "some-certificate",
						Name:            "some-stack-name",
						BOSHAZ:          "some-bosh-az",
					},
					LB: storage.LB{
						Type: "concourse",
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

				terraformManager.ApplyCall.Returns.BBLState = state

				err := command.Execute(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})

		It("returns an error if there is no lb", func() {
			err := command.Execute(storage.State{
				TFState: "some-tf-state",
			})
			Expect(err).To(MatchError(commands.LBNotFound))
		})

		Context("state management", func() {
			It("saves state with no lb type", func() {
				err := command.Execute(storage.State{
					LB: storage.LB{
						Type: "cf",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
					LB: storage.LB{},
				}))
			})
		})

		Context("failure cases", func() {
			It("returns an error when aws credential validator fails to validate", func() {
				credentialValidator.ValidateCall.Returns.Error = errors.New("validate failed")
				err := command.Execute(incomingState)
				Expect(err).To(MatchError("validate failed"))
			})

			Context("when terraform manager fails to apply with terraformManagerError", func() {
				It("return an error", func() {
					terraformManager.ApplyCall.Returns.Error = errors.New("apply failed")
					err := command.Execute(incomingState)
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
					err := command.Execute(incomingState)
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
						err := command.Execute(incomingState)
						Expect(err).To(MatchError("the following errors occurred:\ncannot apply,\nfailed to retrieve bbl state"))
					})
				})
			})

			It("return an error when cloud config manager fails to update", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("update failed")
				err := command.Execute(incomingState)
				Expect(err).To(MatchError("update failed"))
			})

			It("returns an error when the state fails to save lb type", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to save state")}}
				err := command.Execute(incomingState)
				Expect(err).To(MatchError("failed to save state"))
			})
		})
	})
})
