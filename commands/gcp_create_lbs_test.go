package commands_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCPCreateLBs", func() {
	var (
		terraformManager          *fakes.TerraformManager
		cloudConfigManager        *fakes.CloudConfigManager
		stateStore                *fakes.StateStore
		environmentValidator      *fakes.EnvironmentValidator
		terraformExecutorError    *fakes.TerraformExecutorError
		availabilityZoneRetriever *fakes.GCPClient

		bblState    storage.State
		command     commands.GCPCreateLBs
		certPath    string
		keyPath     string
		certificate string
		key         string
	)

	BeforeEach(func() {
		terraformManager = &fakes.TerraformManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		stateStore = &fakes.StateStore{}
		environmentValidator = &fakes.EnvironmentValidator{}
		terraformExecutorError = &fakes.TerraformExecutorError{}
		availabilityZoneRetriever = &fakes.GCPClient{}

		command = commands.NewGCPCreateLBs(terraformManager, cloudConfigManager, stateStore, environmentValidator, availabilityZoneRetriever)

		tempCertFile, err := ioutil.TempFile("", "cert")
		Expect(err).NotTo(HaveOccurred())

		certificate = "some-cert"
		certPath = tempCertFile.Name()
		err = ioutil.WriteFile(certPath, []byte(certificate), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		tempKeyFile, err := ioutil.TempFile("", "key")
		Expect(err).NotTo(HaveOccurred())

		key = "some-key"
		keyPath = tempKeyFile.Name()
		err = ioutil.WriteFile(keyPath, []byte(key), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		bblState = storage.State{
			GCP: storage.GCP{
				Region: "some-region",
			},
			TFState: "some-tfstate",
		}
	})

	AfterEach(func() {
		commands.ResetMarshal()
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			availabilityZoneRetriever.GetZonesCall.Returns.Zones = []string{"z1", "z2", "z3"}
		})
		It("calls terraform manager apply", func() {
			err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{
				LBType:   "cf",
				CertPath: certPath,
				KeyPath:  keyPath,
				Domain:   "some-domain",
			}}, bblState)
			Expect(err).NotTo(HaveOccurred())

			Expect(availabilityZoneRetriever.GetZonesCall.Receives.Region).To(Equal("some-region"))
			Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(storage.State{
				GCP: storage.GCP{
					Zones:  []string{"z1", "z2", "z3"},
					Region: "some-region",
				},
				TFState: "some-tfstate",
				LB: storage.LB{
					Type:   "cf",
					Cert:   certificate,
					Key:    key,
					Domain: "some-domain",
				},
			}))
		})

		It("saves the updated tfstate", func() {
			terraformManager.ApplyCall.Returns.BBLState = storage.State{
				LB: storage.LB{
					Type: "concourse",
				},
				TFState: "some-new-tfstate",
			}

			err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{
				LBType: "concourse",
			}}, storage.State{TFState: "some-old-tfstate"})
			Expect(err).NotTo(HaveOccurred())

			Expect(stateStore.SetCall.CallCount).To(Equal(1))
			Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
				LB: storage.LB{
					Type: "concourse",
				},
				TFState: "some-new-tfstate",
			}))
		})

		It("uploads a new cloud-config to the bosh director", func() {
			terraformManager.ApplyCall.Returns.BBLState = bblState

			err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{
				LBType: "concourse",
			}}, bblState)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(bblState))
		})

		Context("when there is no BOSH director", func() {
			It("does not call the CloudConfigManager", func() {
				terraformManager.ApplyCall.Returns.BBLState.NoDirector = true

				err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{}}, storage.State{
					NoDirector: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})

		Context("failure cases", func() {
			It("returns an error if terraform manager version validator fails", func() {
				terraformManager.ValidateVersionCall.Returns.Error = errors.New("cannot validate version")

				err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{}}, storage.State{})
				Expect(err).To(MatchError("cannot validate version"))
			})

			Context("when environment validator validate returns an error", func() {
				It("returns a DirectorNotReachable error", func() {
					environmentValidator.ValidateCall.Returns.Error = application.DirectorNotReachable

					err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{}}, storage.State{})
					Expect(err).To(MatchError(application.DirectorNotReachable))
				})
			})

			It("returns an error when the availability zone retriever fails to get zones", func() {
				availabilityZoneRetriever.GetZonesCall.Returns.Error = errors.New("failed to get zones")
				err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}}, storage.State{TFState: "some-tf-state"})
				Expect(err).To(MatchError("failed to get zones"))
			})

			It("saves the tf state even if the applier fails", func() {
				terraformExecutorError.TFStateCall.Returns.TFState = "some-updated-tf-state"
				terraformExecutorError.ErrorCall.Returns = "failed to apply"
				expectedError := terraform.NewManagerError(storage.State{
					TFState: "some-tf-state",
				}, terraformExecutorError)
				terraformManager.ApplyCall.Returns.Error = expectedError

				err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}}, storage.State{
					TFState: "some-prev-tf-state",
				})

				Expect(err).To(MatchError("failed to apply"))
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[0].State.TFState).To(Equal("some-updated-tf-state"))
			})

			It("returns an error if terraform manager apply fails with non terraform manager apply error", func() {
				terraformManager.ApplyCall.Returns.Error = errors.New("failed to apply")
				err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: certPath,
					KeyPath:  keyPath,
				}}, storage.State{
					TFState: "some-tf-state",
				})
				Expect(err).To(MatchError("failed to apply"))
				Expect(stateStore.SetCall.CallCount).To(Equal(0))
			})

			It("returns an error when both the applier fails and terraformManagerError.BBLState fails", func() {
				terraformExecutorError.TFStateCall.Returns.Error = errors.New("failed to get tf state")
				terraformExecutorError.ErrorCall.Returns = "failed to apply"
				expectedError := terraform.NewManagerError(bblState, terraformExecutorError)
				terraformManager.ApplyCall.Returns.Error = expectedError

				err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}}, bblState)

				Expect(err).To(MatchError("the following errors occurred:\nfailed to apply,\nfailed to get tf state"))
				Expect(stateStore.SetCall.CallCount).To(Equal(0))
			})

			It("returns an error when both the applier fails and state fails to be set", func() {
				terraformExecutorError.TFStateCall.Returns.TFState = "some-updated-tf-state"
				terraformExecutorError.ErrorCall.Returns = "failed to apply"
				expectedError := terraform.NewManagerError(storage.State{
					LB: storage.LB{
						Type: "concourse",
					},
					TFState: "some-tf-state",
				}, terraformExecutorError)
				terraformManager.ApplyCall.Returns.Error = expectedError

				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("state failed to be set")}}
				err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}}, storage.State{TFState: "some-tf-state"})

				Expect(err).To(MatchError("the following errors occurred:\nfailed to apply,\nstate failed to be set"))
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
					LB:      storage.LB{Type: "concourse"},
					TFState: "some-updated-tf-state",
				}))
			})

			It("returns an error when the state store fails to save the state after applying terraform", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{fakes.SetCallReturn{Error: errors.New("failed to save state")}}

				err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}}, storage.State{TFState: "some-tf-state"})
				Expect(err).To(MatchError("failed to save state"))
			})

			It("returns an error when the cloud config fails to be updated", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("failed to update cloud config")

				err := command.Execute(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}}, storage.State{TFState: "some-tf-state"})
				Expect(err).To(MatchError("failed to update cloud config"))
			})
		})
	})
})
