package terraform_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExecutorApplyError", func() {
	Describe("NewExecutorApplyError", func() {
		It("sets the error passed in", func() {
			err := errors.New("some-error")

			executorApplyError := terraform.NewExecutorApplyError("", err)
			Expect(executorApplyError.Error()).To(Equal(err.Error()))
		})

		It("sets the tf state passed in", func() {
			tfState := "some-tf-state"
			executorApplyError := terraform.NewExecutorApplyError(tfState, nil)

			Expect(executorApplyError.TFState()).To(Equal(tfState))
		})
	})
})
