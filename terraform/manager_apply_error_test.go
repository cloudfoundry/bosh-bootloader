package terraform_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManagerApplyError", func() {
	var (
		executorApplyError *fakes.TerraformExecutorApplyError
	)

	BeforeEach(func() {
		executorApplyError = &fakes.TerraformExecutorApplyError{}
	})

	Describe("Error", func() {
		It("returns the internal error message", func() {
			expectedErrorMessage := "some-error-message"
			executorApplyError.ErrorCall.Returns = expectedErrorMessage

			managerApplyError := terraform.NewManagerApplyError(storage.State{}, executorApplyError)
			Expect(managerApplyError.Error()).To(Equal(expectedErrorMessage))
		})
	})

	Describe("BBLState", func() {
		It("returns the bbl state with additional tf state", func() {
			executorApplyError.TFStateCall.Returns.TFState = "some-tf-state"
			managerApplyError := terraform.NewManagerApplyError(storage.State{
				IAAS: "gcp",
			}, executorApplyError)

			actualBBLState, err := managerApplyError.BBLState()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualBBLState).To(Equal(storage.State{
				IAAS:    "gcp",
				TFState: "some-tf-state",
			}))
		})

		Context("failure cases", func() {
			It("returns an error when ExecutorApplyError.TFState returns an error", func() {
				executorApplyError.TFStateCall.Returns.Error = errors.New("failed to get tf state")
				managerApplyError := terraform.NewManagerApplyError(storage.State{
					IAAS: "gcp",
				}, executorApplyError)

				_, err := managerApplyError.BBLState()
				Expect(err).To(MatchError("failed to get tf state"))
			})
		})
	})
})
