package terraform_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("Manager", func() {
	var (
		executor              *fakes.TerraformExecutor
		templateGenerator     *fakes.TemplateGenerator
		inputGenerator        *fakes.InputGenerator
		outputGenerator       *fakes.OutputGenerator
		logger                *fakes.Logger
		manager               terraform.Manager
		terraformOutputBuffer bytes.Buffer
	)

	BeforeEach(func() {
		executor = &fakes.TerraformExecutor{}
		templateGenerator = &fakes.TemplateGenerator{}
		inputGenerator = &fakes.InputGenerator{}
		outputGenerator = &fakes.OutputGenerator{}
		logger = &fakes.Logger{}

		manager = terraform.NewManager(terraform.NewManagerArgs{
			Executor:              executor,
			TemplateGenerator:     templateGenerator,
			InputGenerator:        inputGenerator,
			OutputGenerator:       outputGenerator,
			TerraformOutputBuffer: &terraformOutputBuffer,
			Logger:                logger,
		})
	})

	Describe("Apply", func() {
		var (
			incomingState    storage.State
			expectedState    storage.State
			expectedTFState  string
			expectedTFOutput string
		)

		BeforeEach(func() {
			incomingState = storage.State{
				IAAS:  "gcp",
				EnvID: "some-env-id",
				GCP: storage.GCP{
					ServiceAccountKey: "some-service-account-key",
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "some-region",
				},
				TFState: "some-tf-state",
				LB: storage.LB{
					Type:   "cf",
					Domain: "some-domain",
				},
			}

			expectedTFState = "some-updated-tf-state"
			executor.ApplyCall.Returns.TFState = expectedTFState

			expectedTFOutput = "some terraform output"

			expectedState = incomingState
			expectedState.TFState = expectedTFState
			expectedState.LatestTFOutput = expectedTFOutput

			templateGenerator.GenerateCall.Returns.Template = "some-gcp-terraform-template"
			inputGenerator.GenerateCall.Returns.Inputs = map[string]string{
				"env_id":        incomingState.EnvID,
				"project_id":    incomingState.GCP.ProjectID,
				"region":        incomingState.GCP.Region,
				"zone":          incomingState.GCP.Zone,
				"credentials":   "some-path",
				"system_domain": incomingState.LB.Domain,
			}
		})

		It("logs steps", func() {
			_, err := manager.Apply(storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Messages).To(ContainSequence([]string{
				"generating terraform template", "applied terraform template",
			}))
		})

		It("returns a state with new tfState and output from executor apply", func() {
			terraformOutputBuffer.Write([]byte(expectedTFOutput))

			state, err := manager.Apply(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(templateGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			Expect(inputGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			Expect(executor.ApplyCall.Receives.Inputs).To(Equal(map[string]string{
				"env_id":        incomingState.EnvID,
				"project_id":    incomingState.GCP.ProjectID,
				"region":        incomingState.GCP.Region,
				"zone":          incomingState.GCP.Zone,
				"credentials":   "some-path",
				"system_domain": incomingState.LB.Domain,
			}))
			Expect(executor.ApplyCall.Receives.TFState).To(Equal("some-tf-state"))
			Expect(executor.ApplyCall.Receives.Template).To(Equal(string("some-gcp-terraform-template")))
			Expect(state).To(Equal(expectedState))
		})

		Context("failure cases", func() {
			Context("when InputGenerator.Generate returns an error", func() {
				BeforeEach(func() {
					inputGenerator.GenerateCall.Returns.Error = errors.New("failed to generate inputs")
				})

				It("bubbles up the error", func() {
					_, err := manager.Apply(incomingState)
					Expect(err).To(MatchError("failed to generate inputs"))
				})
			})

			Context("when Executor.Apply returns a ExecutorError", func() {
				var (
					tempDir       string
					executorError *fakes.TerraformExecutorError
				)

				BeforeEach(func() {
					var err error
					tempDir, err = ioutil.TempDir("", "")
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(filepath.Join(tempDir, "terraform.tfstate"), []byte("updated-tf-state"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					executorError = &fakes.TerraformExecutorError{}
					executor.ApplyCall.Returns.Error = executorError

					terraformOutputBuffer.Write([]byte(expectedTFOutput))
				})

				AfterEach(func() {
					executor.ApplyCall.Returns.Error = nil
				})

				It("returns the bblState with latest terraform output and a ManagerError", func() {
					_, err := manager.Apply(incomingState)

					expectedState := incomingState
					expectedState.LatestTFOutput = expectedTFOutput
					expectedError := terraform.NewManagerError(expectedState, executorError)
					Expect(err).To(MatchError(expectedError))
				})
			})

			Context("when Executor.Apply returns a non-ExecutorError error", func() {
				executorError := errors.New("some-error")

				BeforeEach(func() {
					executor.ApplyCall.Returns.Error = executorError
				})

				AfterEach(func() {
					executor.ApplyCall.Returns.Error = nil
				})

				It("bubbles up the error", func() {
					_, err := manager.Apply(incomingState)
					Expect(err).To(Equal(executorError))
				})
			})
		})
	})

	Describe("Destroy", func() {
		Context("when the bbl state contains a non-empty TFState", func() {
			var (
				incomingState = storage.State{
					EnvID: "some-env-id",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
					LB: storage.LB{
						Type:   "cf",
						Domain: "some-domain",
					},
					TFState: "some-tf-state",
				}
				updatedTFState = "some-updated-tf-state"
			)

			BeforeEach(func() {
				templateGenerator.GenerateCall.Returns.Template = "some-gcp-terraform-template"

				inputGenerator.GenerateCall.Returns.Inputs = map[string]string{
					"env_id":        incomingState.EnvID,
					"project_id":    incomingState.GCP.ProjectID,
					"region":        incomingState.GCP.Region,
					"zone":          incomingState.GCP.Zone,
					"credentials":   "some-path",
					"system_domain": incomingState.LB.Domain,
				}
			})

			It("calls Executor.Destroy with the right arguments", func() {
				_, err := manager.Destroy(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(templateGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

				Expect(inputGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

				Expect(executor.DestroyCall.Receives.Inputs).To(Equal(map[string]string{
					"env_id":        incomingState.EnvID,
					"project_id":    incomingState.GCP.ProjectID,
					"region":        incomingState.GCP.Region,
					"zone":          incomingState.GCP.Zone,
					"credentials":   "some-path",
					"system_domain": incomingState.LB.Domain,
				}))
				Expect(executor.DestroyCall.Receives.Template).To(Equal(templateGenerator.GenerateCall.Returns.Template))
				Expect(executor.DestroyCall.Receives.TFState).To(Equal(incomingState.TFState))
			})

			Context("when Executor.Destroy succeeds", func() {
				BeforeEach(func() {
					executor.DestroyCall.Returns.TFState = updatedTFState
				})

				AfterEach(func() {
					executor.DestroyCall.Returns.TFState = ""
				})

				It("returns the bbl state updated with the TFState returned by Executor.Destroy", func() {
					newBBLState, err := manager.Destroy(incomingState)
					Expect(err).NotTo(HaveOccurred())

					expectedBBLState := incomingState
					expectedBBLState.TFState = updatedTFState
					Expect(newBBLState.TFState).To(Equal(updatedTFState))
					Expect(newBBLState).To(Equal(expectedBBLState))

					Expect(logger.StepCall.Messages).To(ContainSequence([]string{
						"destroying infrastructure", "finished destroying infrastructure",
					}))
				})
			})

			Context("when InputGenerator.Generate returns an error", func() {
				BeforeEach(func() {
					inputGenerator.GenerateCall.Returns.Error = errors.New("failed to generate inputs")
				})

				It("bubbles up the error", func() {
					_, err := manager.Apply(incomingState)
					Expect(err).To(MatchError("failed to generate inputs"))
				})
			})

			Context("when Executor.Destroy returns a ExecutorError", func() {
				var (
					tempDir       string
					executorError *fakes.TerraformExecutorError
				)

				BeforeEach(func() {
					var err error
					tempDir, err = ioutil.TempDir("", "")
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(filepath.Join(tempDir, "terraform.tfstate"), []byte("updated-tf-state"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					executorError = &fakes.TerraformExecutorError{}
					executor.DestroyCall.Returns.Error = executorError
				})

				AfterEach(func() {
					executor.DestroyCall.Returns.Error = nil
				})

				It("returns a ManagerError", func() {
					_, err := manager.Destroy(incomingState)

					expectedError := terraform.NewManagerError(incomingState, executorError)
					Expect(err).To(MatchError(expectedError))
				})
			})

			Context("when Executor.Destroy returns a non-ExecutorError error", func() {
				executorError := errors.New("some-error")

				BeforeEach(func() {
					executor.DestroyCall.Returns.Error = executorError
				})

				AfterEach(func() {
					executor.DestroyCall.Returns.Error = nil
				})

				It("bubbles up the error", func() {
					_, err := manager.Destroy(incomingState)
					Expect(err).To(Equal(executorError))
				})
			})
		})

		Context("when the bbl state contains a non-empty TFState", func() {
			var (
				incomingState = storage.State{
					EnvID: "some-env-id",
				}
			)

			It("returns the bbl state and skips calling executor destroy", func() {
				bblState, err := manager.Destroy(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(bblState).To(Equal(incomingState))
				Expect(executor.DestroyCall.CallCount).To(Equal(0))
			})
		})
	})

	Describe("GetOutputs", func() {
		BeforeEach(func() {
			outputGenerator.GenerateCall.Returns.Outputs = map[string]interface{}{
				"external_ip":        "some-external-ip",
				"network_name":       "some-network-name",
				"subnetwork_name":    "some-subnetwork-name",
				"bosh_open_tag_name": "some-bosh-open-tag-name",
				"internal_tag_name":  "some-internal-tag-name",
				"director_address":   "some-director-address",
			}
		})

		It("returns all terraform outputs except lb related outputs", func() {
			incomingState := storage.State{
				EnvID: "some-env-id",
				GCP: storage.GCP{
					ServiceAccountKey: "some-service-account-key",
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "some-region",
				},
				LB: storage.LB{
					Type:   "some-lb-type",
					Domain: "some-domain",
				},
				TFState: "some-tf-state",
			}

			terraformOutputs, err := manager.GetOutputs(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(outputGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			Expect(terraformOutputs).To(Equal(map[string]interface{}{
				"external_ip":        "some-external-ip",
				"network_name":       "some-network-name",
				"subnetwork_name":    "some-subnetwork-name",
				"bosh_open_tag_name": "some-bosh-open-tag-name",
				"internal_tag_name":  "some-internal-tag-name",
				"director_address":   "some-director-address",
			}))
		})

		Context("failure cases", func() {
			Context("when the output generator fails", func() {
				It("returns the error to the caller", func() {
					outputGenerator.GenerateCall.Returns.Error = errors.New("fail")
					_, err := manager.GetOutputs(storage.State{})
					Expect(err).To(MatchError("fail"))
				})
			})
		})
	})

	Describe("Version", func() {
		BeforeEach(func() {
			executor.VersionCall.Returns.Version = "some-version"
		})

		It("returns a version", func() {
			version, err := manager.Version()
			Expect(err).NotTo(HaveOccurred())

			Expect(executor.VersionCall.CallCount).To(Equal(1))
			Expect(version).To(Equal("some-version"))
		})

		Context("failure cases", func() {
			Context("when version returns an error", func() {
				BeforeEach(func() {
					executor.VersionCall.Returns.Error = errors.New("failed to get version")
				})

				It("returns the error", func() {
					_, err := manager.Version()
					Expect(err).To(MatchError("failed to get version"))
				})
			})
		})
	})

	Describe("ValidateVersion", func() {
		Context("when terraform version is greater than v0.8.5", func() {
			BeforeEach(func() {
				executor.VersionCall.Returns.Version = "0.9.1"
			})

			It("validates the version of terraform and returns no error", func() {
				err := manager.ValidateVersion()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when terraform version is v0.9.0", func() {
			BeforeEach(func() {
				executor.VersionCall.Returns.Version = "0.9.0"
			})

			It("returns a helpful error message", func() {
				err := manager.ValidateVersion()
				Expect(err).To(MatchError("Version 0.9.0 of terraform is incompatible with bbl, please try a later version."))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the terraform installed is less than v0.8.5", func() {
				executor.VersionCall.Returns.Version = "0.8.4"

				err := manager.ValidateVersion()
				Expect(err).To(MatchError("Terraform version must be at least v0.8.5"))
			})

			It("fast fails if the terraform executor fails to get the version", func() {
				executor.VersionCall.Returns.Error = errors.New("cannot get version")

				err := manager.ValidateVersion()
				Expect(err).To(MatchError("cannot get version"))
			})

			It("fast fails when the version cannot be parsed by go-semver", func() {
				executor.VersionCall.Returns.Version = "lol.5.2"

				err := manager.ValidateVersion()
				Expect(err.Error()).To(ContainSubstring("invalid syntax"))
			})
		})
	})
})
