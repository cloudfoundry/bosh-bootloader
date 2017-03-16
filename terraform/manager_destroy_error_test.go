package terraform_test

import (
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManagerDestroyError", func() {

	Describe("NewManagerDestroyError", func() {
		It("sets the executorDestroyError passed in", func() {
			executorDestroyError := &fakes.TerraformExecutorDestroyError{}
			expectedErrorMessage := "some-error-message"
			executorDestroyError.ErrorCall.Returns = expectedErrorMessage

			managerDestroyError := terraform.NewManagerDestroyError(storage.State{}, executorDestroyError)
			Expect(managerDestroyError.Error()).To(Equal(expectedErrorMessage))
		})

		It("sets the bbl state passed in", func() {
			bblState := storage.State{
				IAAS: "gcp",
			}
			managerDestroyError := terraform.NewManagerDestroyError(bblState, terraform.ExecutorDestroyError{})

			Expect(managerDestroyError.BBLState()).To(Equal(bblState))
		})
	})
})
