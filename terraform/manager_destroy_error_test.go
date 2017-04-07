package terraform_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManagerDestroyError", func() {
	var (
		executorDestroyError *fakes.TerraformExecutorDestroyError
	)

	BeforeEach(func() {
		executorDestroyError = &fakes.TerraformExecutorDestroyError{}
	})

	Describe("Error", func() {
		It("returns the internal error message", func() {
			expectedErrorMessage := "some-error-message"
			executorDestroyError.ErrorCall.Returns = expectedErrorMessage

			managerDestroyError := terraform.NewManagerDestroyError(storage.State{}, executorDestroyError)
			Expect(managerDestroyError.Error()).To(Equal(expectedErrorMessage))
		})
	})

	Describe("BBLState", func() {
		It("returns the bbl state with additional tf state", func() {
			executorDestroyError.TFStateCall.Returns.TFState = "some-tf-state"
			managerDestroyError := terraform.NewManagerDestroyError(storage.State{
				IAAS: "gcp",
			}, executorDestroyError)

			actualBBLState, err := managerDestroyError.BBLState()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualBBLState).To(Equal(storage.State{
				IAAS:    "gcp",
				TFState: "some-tf-state",
			}))
		})

		Context("failure cases", func() {
			It("returns an error when ExecutorDestroyError.TFState returns an error", func() {
				executorDestroyError.TFStateCall.Returns.Error = errors.New("failed to get tf state")
				managerDestroyError := terraform.NewManagerDestroyError(storage.State{
					IAAS: "gcp",
				}, executorDestroyError)

				_, err := managerDestroyError.BBLState()
				Expect(err).To(MatchError("failed to get tf state"))
			})
		})
	})
})
