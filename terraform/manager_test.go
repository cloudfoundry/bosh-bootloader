package terraform_test

import (
	"bytes"
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		expectedTFState       string
		expectedTFOutput      string
	)

	BeforeEach(func() {
		executor = &fakes.TerraformExecutor{}
		templateGenerator = &fakes.TemplateGenerator{}
		inputGenerator = &fakes.InputGenerator{}
		outputGenerator = &fakes.OutputGenerator{}
		logger = &fakes.Logger{}

		expectedTFOutput = "some terraform output"
		expectedTFState = "some-updated-tf-state"

		manager = terraform.NewManager(terraform.NewManagerArgs{
			Executor:              executor,
			TemplateGenerator:     templateGenerator,
			InputGenerator:        inputGenerator,
			OutputGenerator:       outputGenerator,
			TerraformOutputBuffer: &terraformOutputBuffer,
			Logger:                logger,
		})
	})

	AfterEach(func() {
		terraformOutputBuffer.Reset()
	})

	Describe("Init", func() {
		var incomingState storage.State

		BeforeEach(func() {
			incomingState = storage.State{
				TFState: "some-tf-state",
			}
			templateGenerator.GenerateCall.Returns.Template = "some-terraform-template"
		})

		It("returns a state with new tfState and output from executor apply", func() {
			err := manager.Init(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(templateGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			Expect(executor.InitCall.CallCount).To(Equal(1))
			Expect(executor.InitCall.Receives.TFState).To(Equal("some-tf-state"))
			Expect(executor.InitCall.Receives.Template).To(Equal(string("some-terraform-template")))

			Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{
				"generating terraform template",
			}))
		})

		Context("when the executor init causes an executor error", func() {
			BeforeEach(func() {
				executor.InitCall.Returns.Error = errors.New("canteloupe")
			})

			It("returns the bblState with latest terraform output and a ManagerError", func() {
				err := manager.Init(incomingState)
				Expect(err).To(MatchError("Executor init: canteloupe"))
			})
		})
	})

	Describe("Apply", func() {
		var (
			incomingState storage.State
			expectedState storage.State
		)

		BeforeEach(func() {
			incomingState = storage.State{
				EnvID:   "some-env-id",
				TFState: "some-tf-state",
			}

			executor.ApplyCall.Returns.TFState = expectedTFState

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

			terraformOutputBuffer.Write([]byte(expectedTFOutput))
		})

		It("returns a state with new tfState and output from executor apply", func() {
			state, err := manager.Apply(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(inputGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			Expect(executor.ApplyCall.Receives.Inputs).To(Equal(map[string]string{
				"env_id":        incomingState.EnvID,
				"project_id":    incomingState.GCP.ProjectID,
				"region":        incomingState.GCP.Region,
				"zone":          incomingState.GCP.Zone,
				"credentials":   "some-path",
				"system_domain": incomingState.LB.Domain,
			}))
			Expect(state).To(Equal(expectedState))

			Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{
				"generating terraform variables",
				"terraform apply",
			}))
		})

		Context("when an error occurs", func() {
			Context("when input generator returns an error", func() {
				BeforeEach(func() {
					inputGenerator.GenerateCall.Returns.Error = errors.New("kiwi")
				})

				It("bubbles up the error", func() {
					_, err := manager.Apply(incomingState)
					Expect(err).To(MatchError("Input generator generate: kiwi"))
				})
			})

			Context("when the applying causes an executor error", func() {
				BeforeEach(func() {
					executor.ApplyCall.Returns.Error = &fakes.TerraformExecutorError{}
				})

				It("returns the bblState with latest terraform output and a ManagerError", func() {
					_, err := manager.Apply(incomingState)
					Expect(err).To(BeAssignableToTypeOf(terraform.ManagerError{}))
				})
			})

			Context("when executor apply returns a non-ExecutorError error", func() {
				BeforeEach(func() {
					executor.ApplyCall.Returns.Error = errors.New("banana")
				})

				It("bubbles up the error", func() {
					_, err := manager.Apply(incomingState)
					Expect(err).To(MatchError("banana"))
				})
			})
		})
	})

	Describe("Destroy", func() {
		Context("when the bbl state contains a non-empty TFState", func() {
			var (
				incomingState storage.State
				expectedState storage.State
			)

			BeforeEach(func() {
				incomingState = storage.State{
					TFState: "some-tf-state",
				}
				executor.DestroyCall.Returns.TFState = expectedTFState

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

			It("calling executor destroy with the right arguments", func() {
				_, err := manager.Destroy(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(templateGenerator.GenerateCall.Receives.State).To(Equal(incomingState))
				Expect(inputGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

				Expect(executor.InitCall.CallCount).To(Equal(1))
				Expect(executor.InitCall.Receives.Template).To(Equal(templateGenerator.GenerateCall.Returns.Template))
				Expect(executor.InitCall.Receives.TFState).To(Equal(incomingState.TFState))
				Expect(executor.DestroyCall.CallCount).To(Equal(1))
				Expect(executor.DestroyCall.Receives.Inputs).To(Equal(map[string]string{
					"env_id":        incomingState.EnvID,
					"project_id":    incomingState.GCP.ProjectID,
					"region":        incomingState.GCP.Region,
					"zone":          incomingState.GCP.Zone,
					"credentials":   "some-path",
					"system_domain": incomingState.LB.Domain,
				}))

				Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{
					"destroying infrastructure",
					"generating terraform template",
					"generating terraform variables",
					"terraform destroy",
					"finished destroying infrastructure",
				}))
			})

			It("returns the bbl state updated with the TFState and output from executor destroy", func() {
				terraformOutputBuffer.Write([]byte(expectedTFOutput))

				newBBLState, err := manager.Destroy(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(newBBLState).To(Equal(expectedState))
			})

			Context("when input generator returns an error", func() {
				BeforeEach(func() {
					inputGenerator.GenerateCall.Returns.Error = errors.New("apple")
				})

				It("bubbles up the error", func() {
					_, err := manager.Destroy(incomingState)
					Expect(err).To(MatchError("Input generator generate: apple"))
				})
			})

			Context("when executor init returns an error", func() {
				BeforeEach(func() {
					executor.InitCall.Returns.Error = errors.New("apple")
				})

				It("bubbles up the error", func() {
					_, err := manager.Destroy(incomingState)
					Expect(err).To(MatchError("Executor init: apple"))
				})
			})

			Context("when Executor.Destroy returns a ExecutorError", func() {
				var executorError *fakes.TerraformExecutorError

				BeforeEach(func() {
					executorError = &fakes.TerraformExecutorError{}
					executor.DestroyCall.Returns.Error = executorError

					terraformOutputBuffer.Write([]byte(expectedTFOutput))
				})

				It("returns a ManagerError", func() {
					_, err := manager.Destroy(incomingState)

					expectedState := incomingState
					expectedState.LatestTFOutput = expectedTFOutput
					expectedError := terraform.NewManagerError(expectedState, executorError)
					Expect(err).To(MatchError(expectedError))
				})
			})

			Context("when Executor.Destroy returns a non-ExecutorError error", func() {
				BeforeEach(func() {
					executor.DestroyCall.Returns.Error = errors.New("pineapple")
				})

				It("bubbles up the error", func() {
					_, err := manager.Destroy(incomingState)
					Expect(err).To(MatchError("Executor destroy: pineapple"))
				})
			})
		})

		Context("when the bbl state contains a non-empty TFState", func() {
			It("returns the bbl state and skips calling executor destroy", func() {
				incomingState := storage.State{EnvID: "some-env-id"}
				bblState, err := manager.Destroy(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(bblState).To(Equal(incomingState))
				Expect(executor.DestroyCall.CallCount).To(Equal(0))
			})
		})
	})

	Describe("GetOutputs", func() {
		BeforeEach(func() {
			outputGenerator.GenerateCall.Returns.Outputs = terraform.Outputs{
				Map: map[string]interface{}{"external_ip": "some-external-ip"},
			}
		})

		It("returns all terraform outputs except lb related outputs", func() {
			incomingState := storage.State{
				IAAS:    "gcp",
				TFState: "some-tf-state",
			}

			terraformOutputs, err := manager.GetOutputs(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(outputGenerator.GenerateCall.Receives.TFState).To(Equal("some-tf-state"))
			Expect(terraformOutputs.Map).To(Equal(map[string]interface{}{
				"external_ip": "some-external-ip",
			}))
		})

		Context("when the output generator fails", func() {
			It("returns the error to the caller", func() {
				outputGenerator.GenerateCall.Returns.Error = errors.New("orange")
				_, err := manager.GetOutputs(storage.State{})
				Expect(err).To(MatchError("orange"))
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

		Context("when executor version returns an error", func() {
			BeforeEach(func() {
				executor.VersionCall.Returns.Error = errors.New("failed to get version")
			})

			It("returns the error", func() {
				_, err := manager.Version()
				Expect(err).To(MatchError("failed to get version"))
			})
		})
	})

	Describe("ValidateVersion", func() {
		Context("when terraform version is greater than the minimum", func() {
			BeforeEach(func() {
				executor.VersionCall.Returns.Version = "9.0.0"
			})

			It("validates the version of terraform and returns no error", func() {
				err := manager.ValidateVersion()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("failure cases", func() {
			Context("when terraform version is less than the minimum", func() {
				It("returns an error", func() {
					executor.VersionCall.Returns.Version = "0.0.1"

					err := manager.ValidateVersion()
					Expect(err).To(MatchError("Terraform version must be at least v0.10.0"))
				})
			})

			Context("when terraform executor fails to get the version", func() {
				It("fast fails", func() {
					executor.VersionCall.Returns.Error = errors.New("cannot get version")

					err := manager.ValidateVersion()
					Expect(err).To(MatchError("cannot get version"))
				})
			})

			Context("when terraform version cannot be parsed by go-semver", func() {
				It("fast fails", func() {
					executor.VersionCall.Returns.Version = "lol.5.2"

					err := manager.ValidateVersion()
					Expect(err.Error()).To(ContainSubstring("invalid syntax"))
				})
			})
		})
	})
})
