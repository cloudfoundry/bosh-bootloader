package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AzureUp", func() {
	var (
		azureUp commands.AzureUp

		azureClient        *fakes.AzureClient
		boshManager        *fakes.BOSHManager
		cloudConfigManager *fakes.CloudConfigManager
		envIDManager       *fakes.EnvIDManager
		logger             *fakes.Logger
		stateStore         *fakes.StateStore
		terraformManager   *fakes.TerraformManager
	)

	BeforeEach(func() {
		azureClient = &fakes.AzureClient{}
		boshManager = &fakes.BOSHManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		envIDManager = &fakes.EnvIDManager{}
		logger = &fakes.Logger{}
		stateStore = &fakes.StateStore{}
		terraformManager = &fakes.TerraformManager{}

		azureUp = commands.NewAzureUp(azureClient, boshManager, cloudConfigManager, envIDManager, logger, stateStore, terraformManager)
	})

	Describe("Execute", func() {
		var (
			state              storage.State
			stateWithEnvID     storage.State
			stateWithTerraform storage.State
			stateWithBOSH      storage.State
		)
		BeforeEach(func() {
			state = storage.State{
				Azure: storage.Azure{
					SubscriptionID: "subscription-id",
					TenantID:       "tenant-id",
					ClientID:       "client-id",
					ClientSecret:   "client-secret",
				},
			}

			stateWithEnvID = state
			stateWithEnvID.EnvID = "some-env-id"
			envIDManager.SyncCall.Returns.State = stateWithEnvID

			stateWithTerraform = stateWithEnvID
			stateWithTerraform.TFState = "some-tf-state"
			terraformManager.ApplyCall.Returns.BBLState = stateWithTerraform

			stateWithBOSH = stateWithTerraform
			stateWithBOSH.BOSH = storage.BOSH{
				DirectorName: "bosh-some-env-id",
				State: map[string]interface{}{
					"new-key": "new-value",
				},
				Variables: "some-variables",
				Manifest:  "some-manifest",
			}
			boshManager.CreateDirectorCall.Returns.State = stateWithBOSH
		})

		It("creates the environment", func() {
			err := azureUp.Execute(commands.AzureUpConfig{}, state)

			Expect(err).NotTo(HaveOccurred())
			Expect(logger.StepCall.CallCount).To(Equal(1))
			Expect(logger.StepCall.Messages).To(Equal([]string{"verifying credentials"}))

			By("validating credentials", func() {
				Expect(azureClient.ValidateCredentialsCall.CallCount).To(Equal(1))

				Expect(azureClient.ValidateCredentialsCall.Receives.SubscriptionID).To(Equal("subscription-id"))
				Expect(azureClient.ValidateCredentialsCall.Receives.TenantID).To(Equal("tenant-id"))
				Expect(azureClient.ValidateCredentialsCall.Receives.ClientID).To(Equal("client-id"))
				Expect(azureClient.ValidateCredentialsCall.Receives.ClientSecret).To(Equal("client-secret"))
			})

			By("syncing the environment id", func() {
				Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
				Expect(envIDManager.SyncCall.Receives.State).To(Equal(state))
				Expect(envIDManager.SyncCall.Receives.Name).To(BeEmpty())
			})

			By("saving the resulting state with the env ID", func() {
				Expect(stateStore.SetCall.CallCount).To(BeNumerically(">=", 1))
				Expect(stateStore.SetCall.Receives[0].State).To(Equal(stateWithEnvID))
			})

			By("creating azure resources via terraform", func() {
				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(stateWithEnvID))
			})

			By("saving the resulting terraform state", func() {
				Expect(stateStore.SetCall.CallCount).To(BeNumerically(">=", 2))
				Expect(stateStore.SetCall.Receives[1].State).To(Equal(stateWithTerraform))
			})

			By("getting the terraform outputs", func() {
				Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
				Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(stateWithTerraform))
			})

			By("creating a bosh", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(1))
				Expect(boshManager.CreateDirectorCall.Receives.State).To(Equal(stateWithTerraform))
			})

			By("saving the bosh state to the state", func() {
				Expect(stateStore.SetCall.CallCount).To(BeNumerically(">=", 3))
				Expect(stateStore.SetCall.Receives[2].State).To(Equal(stateWithBOSH))
			})

			By("updating the cloud config", func() {
				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
				Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(stateWithBOSH))
			})
		})

		Context("given invalid credentials", func() {
			BeforeEach(func() {
				azureClient.ValidateCredentialsCall.Returns.Error = errors.New("invalid credentials")
			})

			It("returns the error", func() {
				err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{
					Azure: storage.Azure{},
				})
				Expect(err).To(MatchError("Error: credentials are invalid"))
				Expect(azureClient.ValidateCredentialsCall.CallCount).To(Equal(1))
			})
		})

		Context("given an environment id", func() {
			var expectedEnvIDState storage.State

			BeforeEach(func() {
				expectedEnvIDState = storage.State{
					Azure: storage.Azure{
						SubscriptionID: "subscription-id",
						TenantID:       "tenant-id",
						ClientID:       "client-id",
						ClientSecret:   "client-secret",
					},
				}
			})

			Context("when called with --name", func() {
				It("creates the environment", func() {
					err := azureUp.Execute(commands.AzureUpConfig{Name: "myenvid"}, expectedEnvIDState)
					Expect(err).NotTo(HaveOccurred())

					Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
					Expect(envIDManager.SyncCall.Receives.State).To(Equal(expectedEnvIDState))
					Expect(envIDManager.SyncCall.Receives.Name).To(Equal("myenvid"))
				})
			})

			It("fast fails if an environment with the same name already exists", func() {
				envIDManager.SyncCall.Returns.Error = errors.New("environment already exists")
				err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{})

				Expect(err).To(MatchError("environment already exists"))
			})
		})

		Context("when state store fails to set after syncing env id", func() {
			It("returns an error", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("set call failed")}}
				err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{})

				Expect(err).To(MatchError("set call failed"))
			})
		})

		Context("terraform manager error handling", func() {
			var terraformManagerError *fakes.TerraformManagerError

			BeforeEach(func() {
				terraformManagerError = &fakes.TerraformManagerError{}
				terraformManagerError.ErrorCall.Returns = "failed to apply"
				terraformManagerError.BBLStateCall.Returns.BBLState = storage.State{
					TFState: "some-updated-tf-state",
				}
			})

			It("saves the terraform state when the applier fails", func() {
				terraformManager.ApplyCall.Returns.Error = terraformManagerError

				err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{})

				Expect(err).To(MatchError("failed to apply"))
				Expect(stateStore.SetCall.CallCount).To(Equal(2))
				Expect(stateStore.SetCall.Receives[1].State.TFState).To(Equal("some-updated-tf-state"))
			})

			It("returns an error when the applier fails and we cannot retrieve the updated bbl state", func() {
				terraformManagerError.BBLStateCall.Returns.Error = errors.New("some-bbl-state-error")
				terraformManager.ApplyCall.Returns.Error = terraformManagerError

				err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{})

				Expect(err).To(MatchError("the following errors occurred:\nfailed to apply,\nsome-bbl-state-error"))
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
			})

			It("returns an error if applier fails with non terraform manager apply error", func() {
				terraformManager.ApplyCall.Returns.Error = errors.New("failed to apply")

				err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{})

				Expect(err).To(MatchError("failed to apply"))
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
			})

			It("returns an error when the terraform manager fails, we can retrieve the updated bbl state, and state fails to be set", func() {
				terraformManagerError.BBLStateCall.Returns.BBLState = stateWithTerraform
				terraformManager.ApplyCall.Returns.Error = terraformManagerError
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("state failed to be set")}}

				err := azureUp.Execute(commands.AzureUpConfig{}, state)

				Expect(err).To(MatchError("the following errors occurred:\nfailed to apply,\nstate failed to be set"))
				Expect(stateStore.SetCall.CallCount).To(Equal(2))
				Expect(stateStore.SetCall.Receives[1].State.TFState).To(Equal("some-tf-state"))
			})
		})

		Context("when state store fails to set after terraform apply", func() {
			It("returns an error when state store fails to set after syncing env id", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {Error: errors.New("set call failed")}}
				err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{})

				Expect(err).To(MatchError("set call failed"))
			})
		})

		Context("when terraform manager get outputs fails", func() {
			It("returns an error", func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("get outputs call failed")
				err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{})

				Expect(err).To(MatchError("get outputs call failed"))
			})
		})

		Context("when bosh manager create director fails", func() {
			It("returns an error", func() {
				boshManager.CreateDirectorCall.Returns.Error = errors.New("create director failed")
				err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{})

				Expect(err).To(MatchError("create director failed"))
			})
		})

		Context("when state store fails to set after create director", func() {
			It("returns an error", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {Error: errors.New("set call failed")}}
				err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{})

				Expect(err).To(MatchError("set call failed"))
			})
		})

		Context("when cloud config manager update fails", func() {
			It("returns an error", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("update failed")
				err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{})

				Expect(err).To(MatchError("update failed"))
			})
		})

		Context("when the no-director flag is provided", func() {
			BeforeEach(func() {
				terraformManager.ApplyCall.Returns.BBLState.NoDirector = true
			})

			It("does not create a bosh director or update cloud config", func() {
				err := azureUp.Execute(commands.AzureUpConfig{
					NoDirector: true,
				}, storage.State{
					Azure: storage.Azure{
						SubscriptionID: "subscription-id",
						TenantID:       "tenant-id",
						ClientID:       "client-id",
						ClientSecret:   "client-secret",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.CallCount).To(Equal(2))
				Expect(stateStore.SetCall.Receives[0].State.NoDirector).To(Equal(true))
				Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(0))
				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})
	})
})
