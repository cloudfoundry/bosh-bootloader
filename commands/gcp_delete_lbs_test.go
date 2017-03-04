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
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("GCPDeleteLBs", func() {
	var (
		cloudConfigManager *fakes.CloudConfigManager
		stateStore         *fakes.StateStore
		logger             *fakes.Logger
		terraformExecutor  *fakes.TerraformExecutor

		command commands.GCPDeleteLBs

		expectedTerraformTemplate string
	)

	Describe("Execute", func() {
		BeforeEach(func() {
			stateStore = &fakes.StateStore{}
			logger = &fakes.Logger{}
			terraformExecutor = &fakes.TerraformExecutor{}
			terraformExecutor.VersionCall.Returns.Version = "0.8.7"
			cloudConfigManager = &fakes.CloudConfigManager{}

			command = commands.NewGCPDeleteLBs(logger, stateStore, terraformExecutor, cloudConfigManager)

			body, err := ioutil.ReadFile("fixtures/terraform_template_no_lb.tf")
			Expect(err).NotTo(HaveOccurred())

			expectedTerraformTemplate = string(body)
		})

		It("updates the cloud config", func() {
			err := command.Execute(storage.State{
				IAAS: "gcp",
				BOSH: storage.BOSH{
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
					DirectorAddress:  "some-director-address",
				},
				GCP: storage.GCP{
					Region: "some-region",
				},
				LB: storage.LB{
					Type: "cf",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(storage.State{
				IAAS: "gcp",
				BOSH: storage.BOSH{
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
					DirectorAddress:  "some-director-address",
				},
				GCP: storage.GCP{
					Region: "some-region",
				},
			}))
		})

		It("runs terraform apply", func() {
			credentials := "some-credentials"
			envID := "some-env-id"
			projectID := "some-project-id"
			zone := "some-zone"
			region := "some-region"
			tfState := "some-tf-state"

			err := command.Execute(storage.State{
				EnvID: envID,
				GCP: storage.GCP{
					ServiceAccountKey: credentials,
					Zone:              zone,
					Region:            region,
					ProjectID:         projectID,
				},
				TFState: tfState,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformExecutor.ApplyCall.CallCount).To(Equal(1))
			Expect(terraformExecutor.ApplyCall.Receives.Credentials).To(Equal(credentials))
			Expect(terraformExecutor.ApplyCall.Receives.EnvID).To(Equal(envID))
			Expect(terraformExecutor.ApplyCall.Receives.ProjectID).To(Equal(projectID))
			Expect(terraformExecutor.ApplyCall.Receives.Zone).To(Equal(zone))
			Expect(terraformExecutor.ApplyCall.Receives.Region).To(Equal(region))
			Expect(terraformExecutor.ApplyCall.Receives.Template).To(Equal(expectedTerraformTemplate))
			Expect(terraformExecutor.ApplyCall.Receives.TFState).To(Equal(tfState))

			Expect(logger.StepCall.Messages).To(ContainSequence([]string{
				"generating terraform template", "finished applying terraform template",
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
				Expect(stateStore.SetCall.Receives.State.Stack.LBType).To(Equal(""))
			})

			It("saves the tf state", func() {
				terraformExecutor.ApplyCall.Returns.TFState = "some-tf-state"
				err := command.Execute(storage.State{
					IAAS: "gcp",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives.State.TFState).To(Equal("some-tf-state"))
			})

			It("saves the tf state even if the applier failed", func() {
				expectedError := terraform.NewTerraformApplyError("some-tf-state", errors.New("failed to apply"))
				terraformExecutor.ApplyCall.Returns.Error = expectedError
				err := command.Execute(storage.State{
					IAAS: "gcp",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).To(MatchError(expectedError))
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives.State.TFState).To(Equal("some-tf-state"))
			})
		})

		Context("failure cases", func() {
			It("fast fails if the terraform installed is less than v0.8.5", func() {
				terraformExecutor.VersionCall.Returns.Version = "0.8.4"

				err := command.Execute(storage.State{
					IAAS: "gcp",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})

				Expect(err).To(MatchError("Terraform version must be at least v0.8.5"))
			})
			It("returns an error if applier fails with non terraform apply error", func() {
				terraformExecutor.ApplyCall.Returns.Error = errors.New("failed to apply")
				err := command.Execute(storage.State{
					IAAS: "gcp",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).To(MatchError("failed to apply"))
				Expect(stateStore.SetCall.CallCount).To(Equal(0))
			})

			Context("when terraform applier fails and if fails to save the state", func() {
				It("returns an error with both errors that occurred", func() {
					expectedError := terraform.NewTerraformApplyError("some-tf-state", errors.New("failed to apply"))
					terraformExecutor.ApplyCall.Returns.Error = expectedError
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
