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
		expectedTFOutput      string
	)

	BeforeEach(func() {
		executor = &fakes.TerraformExecutor{}
		templateGenerator = &fakes.TemplateGenerator{}
		inputGenerator = &fakes.InputGenerator{}
		outputGenerator = &fakes.OutputGenerator{}
		logger = &fakes.Logger{}

		expectedTFOutput = "some terraform output"

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

	Describe("IsInitialized", func() {
		It("returns the output of executor.ShouldInit", func() {
			executor.IsInitializedCall.Returns.IsInitialized = true
			Expect(manager.IsInitialized()).To(BeTrue())

			executor.IsInitializedCall.Returns.IsInitialized = false
			Expect(manager.IsInitialized()).To(BeFalse())
		})
	})

	Describe("Init", func() {
		var incomingState storage.State

		BeforeEach(func() {
			inputGenerator.GenerateCall.Returns.Inputs = map[string]interface{}{
				"env_id":        incomingState.EnvID,
				"project_id":    incomingState.GCP.ProjectID,
				"region":        incomingState.GCP.Region,
				"zone":          incomingState.GCP.Zone,
				"credentials":   "some-path",
				"system_domain": incomingState.LB.Domain,
			}

			incomingState = storage.State{
				TFState: "some-tf-state",
			}
			templateGenerator.GenerateCall.Returns.Template = "some-terraform-template"
		})

		It("returns a state with new tfState and output from executor apply", func() {
			err := manager.Init(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(templateGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			Expect(inputGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			Expect(executor.InitCall.CallCount).To(Equal(1))
			Expect(executor.InitCall.Receives.Template).To(Equal(string("some-terraform-template")))
			Expect(executor.InitCall.Receives.Inputs).To(Equal(map[string]interface{}{
				"env_id":        incomingState.EnvID,
				"project_id":    incomingState.GCP.ProjectID,
				"region":        incomingState.GCP.Region,
				"zone":          incomingState.GCP.Zone,
				"credentials":   "some-path",
				"system_domain": incomingState.LB.Domain,
			}))

			Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{
				"generating terraform template",
				"generating terraform variables",
			}))
		})

		Context("failure cases", func() {
			Context("when input generator returns an error", func() {
				BeforeEach(func() {
					inputGenerator.GenerateCall.Returns.Error = errors.New("kiwi")
				})

				It("bubbles up the error", func() {
					err := manager.Init(incomingState)
					Expect(err).To(MatchError("Input generator generate: kiwi"))
				})
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
	})

	Describe("Apply", func() {
		var (
			incomingState storage.State
			expectedState storage.State
		)

		BeforeEach(func() {
			incomingState = storage.State{
				EnvID: "some-env-id",
			}

			expectedState = incomingState
			expectedState.LatestTFOutput = expectedTFOutput

			templateGenerator.GenerateCall.Returns.Template = "some-gcp-terraform-template"
			terraformOutputBuffer.Write([]byte(expectedTFOutput))
		})

		It("returns a state with new tfState and output from executor apply", func() {
			state, err := manager.Apply(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(state).To(Equal(expectedState))

			Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{
				"terraform apply",
			}))
		})

		Context("when an error occurs", func() {
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
		var (
			incomingState storage.State
			expectedState storage.State
		)

		BeforeEach(func() {
			incomingState = storage.State{}

			expectedState = incomingState
			expectedState.LatestTFOutput = expectedTFOutput

			inputGenerator.GenerateCall.Returns.Inputs = map[string]interface{}{
				"env_id":        incomingState.EnvID,
				"project_id":    incomingState.GCP.ProjectID,
				"region":        incomingState.GCP.Region,
				"zone":          incomingState.GCP.Zone,
				"credentials":   "some-path",
				"system_domain": incomingState.LB.Domain,
			}

			terraformOutputBuffer.Write([]byte(expectedTFOutput))
		})

		It("calling executor destroy with the right arguments", func() {
			newBBLState, err := manager.Destroy(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(inputGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			Expect(executor.DestroyCall.CallCount).To(Equal(1))
			Expect(executor.DestroyCall.Receives.Inputs).To(Equal(map[string]interface{}{
				"env_id":        incomingState.EnvID,
				"project_id":    incomingState.GCP.ProjectID,
				"region":        incomingState.GCP.Region,
				"zone":          incomingState.GCP.Zone,
				"credentials":   "some-path",
				"system_domain": incomingState.LB.Domain,
			}))

			Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{
				"destroying infrastructure",
				"generating terraform variables",
				"terraform destroy",
				"finished destroying infrastructure",
			}))

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

		Context("when Executor.Destroy returns a ExecutorError", func() {
			var executorError *fakes.TerraformExecutorError

			BeforeEach(func() {
				executorError = &fakes.TerraformExecutorError{}
				executor.DestroyCall.Returns.Error = executorError
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

	Describe("GetOutputs", func() {
		BeforeEach(func() {
			outputGenerator.GenerateCall.Returns.Outputs = terraform.Outputs{
				Map: map[string]interface{}{"external_ip": "some-external-ip"},
			}
		})

		It("returns all terraform outputs except lb related outputs", func() {
			terraformOutputs, err := manager.GetOutputs()
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformOutputs.Map).To(Equal(map[string]interface{}{
				"external_ip": "some-external-ip",
			}))
		})

		Context("when the output generator fails", func() {
			It("returns the error to the caller", func() {
				outputGenerator.GenerateCall.Returns.Error = errors.New("orange")
				_, err := manager.GetOutputs()
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
