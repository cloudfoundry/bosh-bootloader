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
		logger                *fakes.Logger
		manager               terraform.Manager
		terraformOutputBuffer bytes.Buffer
		expectedTFOutput      string
	)

	BeforeEach(func() {
		executor = &fakes.TerraformExecutor{}
		templateGenerator = &fakes.TemplateGenerator{}
		inputGenerator = &fakes.InputGenerator{}
		logger = &fakes.Logger{}

		expectedTFOutput = "some terraform output"

		manager = terraform.NewManager(executor, templateGenerator, inputGenerator, &terraformOutputBuffer, logger)
	})

	AfterEach(func() {
		terraformOutputBuffer.Reset()
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
			Expect(executor.SetupCall.CallCount).To(Equal(1))
			Expect(executor.SetupCall.Receives.Template).To(Equal(string("some-terraform-template")))
			Expect(executor.SetupCall.Receives.Inputs).To(Equal(map[string]interface{}{
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
				"terraform init",
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

			Context("when the executor setup causes an executor error", func() {
				BeforeEach(func() {
					executor.SetupCall.Returns.Error = errors.New("canteloupe")
				})

				It("returns the bbl state with latest terraform output and a ManagerError", func() {
					err := manager.Init(incomingState)
					Expect(err).To(MatchError("Executor setup: canteloupe"))
				})
			})
		})
	})

	Describe("Apply", func() {
		var (
			incomingState storage.State
			expectedState storage.State
			credentials   map[string]string
		)

		BeforeEach(func() {
			incomingState = storage.State{
				EnvID: "some-env-id",
			}
			credentials = map[string]string{
				"some-credential": "some-credential-value",
			}

			expectedState = incomingState
			expectedState.LatestTFOutput = expectedTFOutput

			templateGenerator.GenerateCall.Returns.Template = "some-gcp-terraform-template"
			inputGenerator.CredentialsCall.Returns.Credentials = credentials
			terraformOutputBuffer.Write([]byte(expectedTFOutput))
		})

		It("returns a state with new tfState and output from executor apply", func() {
			state, err := manager.Apply(incomingState)

			Expect(executor.ApplyCall.Receives.Credentials).To(Equal(credentials))
			Expect(err).NotTo(HaveOccurred())
			Expect(state).To(Equal(expectedState))
			Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{
				"terraform apply",
			}))
		})

		Context("when executor apply fails", func() {
			BeforeEach(func() {
				executor.ApplyCall.Returns.Error = errors.New("grape")
				incomingState.LatestTFOutput = "some terraform output"
			})

			It("returns the bbl state and the error", func() {
				state, err := manager.Apply(incomingState)
				Expect(err).To(MatchError("Executor apply: grape"))
				Expect(state.LatestTFOutput).To(Equal(incomingState.LatestTFOutput))
			})
		})
	})

	Describe("Destroy", func() {
		var (
			incomingState storage.State
			expectedState storage.State
			credentials   map[string]string
		)

		BeforeEach(func() {
			incomingState = storage.State{}

			expectedState = incomingState
			expectedState.LatestTFOutput = expectedTFOutput

			terraformOutputBuffer.Write([]byte(expectedTFOutput))
			credentials = map[string]string{
				"some-credential": "some-credential-value",
			}
			inputGenerator.CredentialsCall.Returns.Credentials = credentials
		})

		It("calling executor destroy with the right arguments", func() {
			newBBLState, err := manager.Destroy(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(inputGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			Expect(executor.DestroyCall.CallCount).To(Equal(1))
			Expect(executor.DestroyCall.Receives.Credentials).To(Equal(credentials))
			Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{
				"terraform destroy",
				"finished destroying infrastructure",
			}))

			Expect(newBBLState).To(Equal(expectedState))
		})

		Context("when executor destroy fails", func() {
			BeforeEach(func() {
				executor.DestroyCall.Returns.Error = errors.New("grape")
				incomingState.LatestTFOutput = "some terraform output"
			})

			It("returns the current bbl state and the error", func() {
				state, err := manager.Destroy(incomingState)
				Expect(err).To(MatchError("Executor destroy: grape"))
				Expect(state.LatestTFOutput).To(Equal(incomingState.LatestTFOutput))
			})
		})
	})

	Describe("GetOutputs", func() {
		BeforeEach(func() {
			executor.OutputsCall.Returns.Outputs = map[string]interface{}{"external_ip": "some-external-ip"}
		})

		It("returns all terraform outputs except lb related outputs", func() {
			terraformOutputs, err := manager.GetOutputs()
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformOutputs.Map).To(Equal(map[string]interface{}{
				"external_ip": "some-external-ip",
			}))
		})

		Context("when the executor outputs fails", func() {
			It("returns the error", func() {
				executor.OutputsCall.Returns.Error = errors.New("orange")

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
					Expect(err).To(MatchError("Terraform version must be at least v0.11.0"))
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
