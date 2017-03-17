package terraform_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var terraformTemplate = `variable "project_id" {
	type = "string"
}

variable "region" {
	type = "string"
}

variable "zone" {
	type = "string"
}

variable "env_id" {
	type = "string"
}

variable "credentials" {
	type = "string"
}

provider "google" {
	credentials = "${file("${var.credentials}")}"
	project = "${var.project_id}"
	region = "${var.region}"
}
`

var _ = Describe("Manager", func() {
	var (
		executor *fakes.TerraformExecutor
		logger   *fakes.Logger
		manager  terraform.Manager
	)

	BeforeEach(func() {
		executor = &fakes.TerraformExecutor{}
		logger = &fakes.Logger{}
		manager = terraform.NewManager(executor, logger)
	})

	Describe("Destroy", func() {
		Context("when the bbl state contains a non-empty TFState", func() {
			var (
				originalBBLState = storage.State{
					EnvID: "some-env-id",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
					TFState: "some-tf-state",
				}
				updatedTFState = "some-updated-tf-state"
			)

			It("calls Executor.Destroy with the right arguments", func() {
				_, err := manager.Destroy(originalBBLState)
				Expect(err).NotTo(HaveOccurred())

				Expect(executor.DestroyCall.Receives.Credentials).To(Equal(originalBBLState.GCP.ServiceAccountKey))
				Expect(executor.DestroyCall.Receives.EnvID).To(Equal(originalBBLState.EnvID))
				Expect(executor.DestroyCall.Receives.ProjectID).To(Equal(originalBBLState.GCP.ProjectID))
				Expect(executor.DestroyCall.Receives.Zone).To(Equal(originalBBLState.GCP.Zone))
				Expect(executor.DestroyCall.Receives.Region).To(Equal(originalBBLState.GCP.Region))
				Expect(executor.DestroyCall.Receives.Template).To(Equal(terraformTemplate))
				Expect(executor.DestroyCall.Receives.TFState).To(Equal(originalBBLState.TFState))
			})

			Context("when Executor.Destroy succeeds", func() {
				BeforeEach(func() {
					executor.DestroyCall.Returns.TFState = updatedTFState
				})

				AfterEach(func() {
					executor.DestroyCall.Returns.TFState = ""
				})

				It("returns the bbl state updated with the TFState returned by Executor.Destroy", func() {
					newBBLState, err := manager.Destroy(originalBBLState)
					Expect(err).NotTo(HaveOccurred())

					expectedBBLState := originalBBLState
					expectedBBLState.TFState = updatedTFState
					Expect(newBBLState.TFState).To(Equal(updatedTFState))
					Expect(newBBLState).To(Equal(expectedBBLState))
				})
			})

			Context("when Executor.Destroy returns a ExecutorDestroyError", func() {
				executorError := terraform.NewExecutorDestroyError(updatedTFState, errors.New("some-error"))

				BeforeEach(func() {
					executor.DestroyCall.Returns.Error = executorError
				})

				AfterEach(func() {
					executor.DestroyCall.Returns.Error = nil
				})

				It("returns a ManagerDestroyError", func() {
					_, err := manager.Destroy(originalBBLState)

					expectedBBLState := originalBBLState
					expectedBBLState.TFState = updatedTFState
					expectedError := terraform.NewManagerDestroyError(expectedBBLState, executorError)
					Expect(err).To(MatchError(expectedError))
				})
			})

			Context("when Executor.Destroy returns a non-ExecutorDestroyError error", func() {
				executorError := errors.New("some-error")

				BeforeEach(func() {
					executor.DestroyCall.Returns.Error = executorError
				})

				AfterEach(func() {
					executor.DestroyCall.Returns.Error = nil
				})

				It("bubbles up the error", func() {
					_, err := manager.Destroy(originalBBLState)
					Expect(err).To(Equal(executorError))
				})
			})
		})

		Context("when the bbl state contains a non-empty TFState", func() {
			var (
				originalBBLState = storage.State{
					EnvID: "some-env-id",
				}
			)

			It("returns the bbl state and skips calling executor destroy", func() {
				bblState, err := manager.Destroy(originalBBLState)
				Expect(err).NotTo(HaveOccurred())

				Expect(bblState).To(Equal(originalBBLState))
				Expect(executor.DestroyCall.CallCount).To(Equal(0))
			})
		})
	})
})
