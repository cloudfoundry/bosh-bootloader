package commands_test

import (
	"errors"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCPDeleteLBs", func() {
	var (
		cloudConfigManager     *fakes.CloudConfigManager
		stateStore             *fakes.StateStore
		environmentValidator   *fakes.EnvironmentValidator
		terraformManager       *fakes.TerraformManager
		terraformExecutorError *fakes.TerraformExecutorError

		command commands.GCPDeleteLBs

		expectedTerraformTemplate string
	)

	Describe("Execute", func() {
		BeforeEach(func() {
			stateStore = &fakes.StateStore{}
			environmentValidator = &fakes.EnvironmentValidator{}
			terraformManager = &fakes.TerraformManager{}
			cloudConfigManager = &fakes.CloudConfigManager{}
			terraformExecutorError = &fakes.TerraformExecutorError{}

			command = commands.NewGCPDeleteLBs(stateStore, environmentValidator, terraformManager, cloudConfigManager)

			body, err := ioutil.ReadFile("fixtures/terraform_template_no_lb.tf")
			Expect(err).NotTo(HaveOccurred())

			expectedTerraformTemplate = string(body)
		})

		Context("when bbl has a bosh director", func() {
			It("updates the cloud config", func() {
				state := storage.State{
					IAAS: "gcp",
					BOSH: storage.BOSH{
						DirectorUsername: "some-director-username",
					},
					GCP: storage.GCP{
						Region: "some-region",
					},
					LB: storage.LB{
						Type: "cf",
					},
				}
				err := command.Execute(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(environmentValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(environmentValidator.ValidateCall.Receives.State).To(Equal(state))

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
				Expect(cloudConfigManager.UpdateCall.Receives.State.LB.Type).To(Equal(""))
			})
		})

		Context("when bbl does not have a bosh director", func() {
			It("does not update the cloud config", func() {
				err := command.Execute(storage.State{
					IAAS:       "gcp",
					NoDirector: true,
					GCP: storage.GCP{
						Region: "some-region",
					},
					LB: storage.LB{
						Type: "cf",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})

		It("runs terraform apply", func() {
			state := storage.State{
				GCP: storage.GCP{
					ProjectID: "some-project-id",
				},
				LB: storage.LB{
					Type: "concourse",
				},
				TFState: "some-tf-state",
			}

			err := command.Execute(state)
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
			Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(storage.State{
				GCP: storage.GCP{
					ProjectID: "some-project-id",
				},
				TFState: "some-tf-state",
			}))
		})

		Context("state manipulation", func() {
			It("removes the lb from the state", func() {
				err := command.Execute(storage.State{
					IAAS: "gcp",
					LB: storage.LB{
						Type: "concourse",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[0].State.Stack.LBType).To(Equal(""))
			})

			It("saves the tf state", func() {
				terraformManager.ApplyCall.Returns.BBLState = storage.State{
					IAAS: "gcp",
				}

				err := command.Execute(storage.State{
					IAAS: "gcp",
					LB: storage.LB{
						Type: "concourse",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
					IAAS: "gcp",
				}))
			})

			It("saves the tf state even if the applier failed", func() {
				terraformExecutorError.TFStateCall.Returns.TFState = "some-updated-tf-state"
				terraformExecutorError.ErrorCall.Returns = "failed to apply"
				expectedError := terraform.NewManagerError(storage.State{
					TFState: "some-tf-state",
				}, terraformExecutorError)
				terraformManager.ApplyCall.Returns.Error = expectedError

				err := command.Execute(storage.State{
					IAAS: "gcp",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).To(MatchError(expectedError))

				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
					TFState: "some-updated-tf-state",
				}))
			})
		})

		Context("failure cases", func() {
			It("fast fails if the environment validator fails", func() {
				environmentValidator.ValidateCall.Returns.Error = errors.New("invalid")

				err := command.Execute(storage.State{})
				Expect(err).To(MatchError("invalid"))
			})

			It("fast fails if the terraform version is invalid", func() {
				terraformManager.ValidateVersionCall.Returns.Error = errors.New("invalid")

				err := command.Execute(storage.State{
					IAAS: "gcp",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).To(MatchError("invalid"))
			})

			It("returns an error if applier fails with non terraform apply error", func() {
				terraformManager.ApplyCall.Returns.Error = errors.New("failed to apply")

				err := command.Execute(storage.State{
					IAAS: "gcp",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).To(MatchError("failed to apply"))

				Expect(stateStore.SetCall.CallCount).To(Equal(0))
			})

			Context("when terraform applier fails and it fails to save the state", func() {
				BeforeEach(func() {
					terraformExecutorError.TFStateCall.Returns.TFState = "some-updated-tf-state"
					terraformExecutorError.ErrorCall.Returns = "failed to apply"
				})

				It("returns an error with both errors that occurred", func() {
					expectedError := terraform.NewManagerError(storage.State{
						TFState: "some-tf-state",
					}, terraformExecutorError)
					terraformManager.ApplyCall.Returns.Error = expectedError

					stateStore.SetCall.Returns = []fakes.SetCallReturn{
						{errors.New("failed to set state")},
					}

					err := command.Execute(storage.State{
						IAAS: "gcp",
						Stack: storage.Stack{
							LBType: "concourse",
						},
					})
					Expect(err).To(MatchError("the following errors occurred:\nfailed to apply,\nfailed to set state"))
				})
			})

			Context("when terraform applier fails and it fails to get the bbl state", func() {
				BeforeEach(func() {
					terraformExecutorError.TFStateCall.Returns.Error = errors.New("failed to get tf state")
					terraformExecutorError.ErrorCall.Returns = "failed to apply"
				})

				It("returns an error with both errors that occurred", func() {
					expectedError := terraform.NewManagerError(storage.State{
						TFState: "some-tf-state",
					}, terraformExecutorError)
					terraformManager.ApplyCall.Returns.Error = expectedError

					stateStore.SetCall.Returns = []fakes.SetCallReturn{
						{errors.New("failed to set state")},
					}

					err := command.Execute(storage.State{
						IAAS: "gcp",
						Stack: storage.Stack{
							LBType: "concourse",
						},
					})
					Expect(err).To(MatchError("the following errors occurred:\nfailed to apply,\nfailed to get tf state"))
				})
			})

			It("returns an error when updating cloud config fails", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("updating cloud config failed")

				err := command.Execute(storage.State{
					IAAS: "gcp",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).To(MatchError("updating cloud config failed"))
			})

			It("returns an error when setting the state store fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{
					{errors.New("failed to set state")},
				}

				err := command.Execute(storage.State{
					IAAS: "gcp",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).To(MatchError("failed to set state"))
			})
		})
	})
})
