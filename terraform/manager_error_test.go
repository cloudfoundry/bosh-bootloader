package terraform_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManagerError", func() {
	var (
		executorError *fakes.TerraformExecutorError
	)

	BeforeEach(func() {
		executorError = &fakes.TerraformExecutorError{}
	})

	Describe("Error", func() {
		It("returns the internal error message", func() {
			expectedErrorMessage := "some-error-message"
			executorError.ErrorCall.Returns = expectedErrorMessage

			managerError := terraform.NewManagerError(storage.State{}, executorError)
			Expect(managerError.Error()).To(Equal(expectedErrorMessage))
		})
	})

	Describe("BBLState", func() {
		It("returns the bbl state with additional tf state", func() {
			executorError.TFStateCall.Returns.TFState = "some-tf-state"
			managerError := terraform.NewManagerError(storage.State{
				IAAS: "gcp",
			}, executorError)

			actualBBLState, err := managerError.BBLState()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualBBLState).To(Equal(storage.State{
				IAAS: "gcp",
			}))
		})

		Context("failure cases", func() {
			Context("when tfstate call returns an error", func() {
				BeforeEach(func() {
					executorError.TFStateCall.Returns.Error = errors.New("failed to get tf state")
				})

				It("returns an error", func() {
					managerError := terraform.NewManagerError(storage.State{
						IAAS: "gcp",
					}, executorError)

					_, err := managerError.BBLState()
					Expect(err).To(MatchError("failed to get tf state"))
				})
			})
		})
	})
})
