package commands_test

import (
	"errors"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("GCPDeleteLBs", func() {
	var (
		cloudConfigGenerator    *fakes.GCPCloudConfigGenerator
		terraformOutputProvider *fakes.TerraformOutputProvider
		stateStore              *fakes.StateStore
		zones                   *fakes.Zones
		logger                  *fakes.Logger
		boshClientProvider      *fakes.BOSHClientProvider
		boshClient              *fakes.BOSHClient
		terraformExecutor       *fakes.TerraformExecutor

		command commands.GCPDeleteLBs

		expectedTerraformTemplate string
	)

	Describe("Execute", func() {
		BeforeEach(func() {
			cloudConfigGenerator = &fakes.GCPCloudConfigGenerator{}
			terraformOutputProvider = &fakes.TerraformOutputProvider{}
			stateStore = &fakes.StateStore{}
			zones = &fakes.Zones{}
			logger = &fakes.Logger{}
			boshClient = &fakes.BOSHClient{}
			boshClientProvider = &fakes.BOSHClientProvider{}
			boshClientProvider.ClientCall.Returns.Client = boshClient
			terraformExecutor = &fakes.TerraformExecutor{}

			command = commands.NewGCPDeleteLBs(terraformOutputProvider, cloudConfigGenerator, zones, logger, boshClientProvider, stateStore, terraformExecutor)

			body, err := ioutil.ReadFile("fixtures/terraform_template_no_lb.tf")
			Expect(err).NotTo(HaveOccurred())

			expectedTerraformTemplate = string(body)
		})

		It("updates the cloud config", func() {
			terraformOutputProvider.GetCall.Returns.Outputs = terraform.Outputs{
				NetworkName:    "some-network-name",
				SubnetworkName: "some-subnetwork-name",
				InternalTag:    "some-internal-tag",
			}
			zones.GetCall.Returns.Zones = []string{"region1", "region2"}

			expectedCloudConfig := gcp.CloudConfig{
				AZs: []gcp.AZ{
					{
						Name: "region1",
					},
					{
						Name: "region2",
					},
				},
			}
			cloudConfigGenerator.GenerateCall.Returns.CloudConfig = expectedCloudConfig

			expectedCloudConfigYAML, err := yaml.Marshal(expectedCloudConfig)
			Expect(err).NotTo(HaveOccurred())

			err = command.Execute(storage.State{
				IAAS: "gcp",
				BOSH: storage.BOSH{
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
					DirectorAddress:  "some-director-address",
				},
				GCP: storage.GCP{
					Region: "some-region",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(zones.GetCall.CallCount).To(Equal(1))
			Expect(zones.GetCall.Receives.Region).To(Equal("some-region"))

			Expect(cloudConfigGenerator.GenerateCall.CallCount).To(Equal(1))
			Expect(cloudConfigGenerator.GenerateCall.Receives.CloudConfigInput).To(Equal(gcp.CloudConfigInput{
				AZs:            []string{"region1", "region2"},
				Tags:           []string{"some-internal-tag"},
				NetworkName:    "some-network-name",
				SubnetworkName: "some-subnetwork-name",
			}))

			Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))

			Expect(boshClient.UpdateCloudConfigCall.CallCount).To(Equal(1))
			Expect(boshClient.UpdateCloudConfigCall.Receives.Yaml).To(MatchYAML(expectedCloudConfigYAML))

			Expect(logger.StepCall.Messages).To(ContainSequence([]string{
				"generating cloud config", "applying cloud config",
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

			It("returns an error when terraform output provider fails", func() {
				terraformOutputProvider.GetCall.Returns.Error = errors.New("failed to return terraform output")
				err := command.Execute(storage.State{
					IAAS: "gcp",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).To(MatchError("failed to return terraform output"))
			})

			Context("when marshaling the cloud config fails", func() {
				BeforeEach(func() {
					commands.SetMarshal(func(interface{}) ([]byte, error) {
						return nil, errors.New("failed to marshal")
					})
				})

				AfterEach(func() {
					commands.ResetMarshal()
				})

				It("returns an error", func() {
					err := command.Execute(storage.State{
						IAAS: "gcp",
						Stack: storage.Stack{
							LBType: "concourse",
						},
					})
					Expect(err).To(MatchError("failed to marshal"))
				})
			})

			It("returns an error when updating cloud config fails", func() {
				boshClient.UpdateCloudConfigCall.Returns.Error = errors.New("updating cloud config failed")
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
