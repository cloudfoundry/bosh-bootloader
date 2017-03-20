package terraform_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExecutorDestroyError", func() {
	Describe("NewExecutorDestroyError", func() {
		It("sets the error passed in", func() {
			err := errors.New("some-error")

			executorDestroyError := terraform.NewExecutorDestroyError("", err)
			Expect(executorDestroyError.Error()).To(Equal(err.Error()))
		})

		It("sets the tf state passed in", func() {
			tfState := "some-tf-state"
			executorDestroyError := terraform.NewExecutorDestroyError(tfState, nil)

			Expect(executorDestroyError.TFState()).To(Equal(tfState))
		})
	})
})
