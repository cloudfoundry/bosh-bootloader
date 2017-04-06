package terraform_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExecutorApplyError", func() {
	Describe("Error", func() {
		It("returns just the internal error message when debug is true", func() {
			err := errors.New("some-error")
			executorApplyError := terraform.NewExecutorApplyError("", err, true)

			Expect(executorApplyError.Error()).To(Equal(err.Error()))
		})

		It("returns the internal error message and mentions the --debug flag when debug is false", func() {
			err := errors.New("some-error")
			executorApplyError := terraform.NewExecutorApplyError("", err, false)

			Expect(executorApplyError.Error()).To(Equal(fmt.Sprintf("%s\n%s", err.Error(), "use --debug for additional debug output")))
		})
	})

	Describe("TFState", func() {
		It("returns the tfState", func() {
			tfState := "some-tf-state"
			executorApplyError := terraform.NewExecutorApplyError(tfState, nil, true)

			Expect(executorApplyError.TFState()).To(Equal(tfState))
		})
	})
})
