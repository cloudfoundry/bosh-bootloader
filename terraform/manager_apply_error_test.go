package terraform_test

import (
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManagerApplyError", func() {

	Describe("NewManagerApplyError", func() {
		It("sets the executorApplyError passed in", func() {
			executorApplyError := &fakes.TerraformExecutorApplyError{}
			expectedErrorMessage := "some-error-message"
			executorApplyError.ErrorCall.Returns = expectedErrorMessage

			managerApplyError := terraform.NewManagerApplyError(storage.State{}, executorApplyError)
			Expect(managerApplyError.Error()).To(Equal(expectedErrorMessage))
		})

		It("sets the bbl state passed in", func() {
			bblState := storage.State{
				IAAS: "gcp",
			}
			managerApplyError := terraform.NewManagerApplyError(bblState, terraform.ExecutorApplyError{})

			Expect(managerApplyError.BBLState()).To(Equal(bblState))
		})
	})
})
