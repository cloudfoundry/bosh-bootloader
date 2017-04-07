package terraform_test

import (
	"errors"
	"fmt"
	"io/ioutil"

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
		var (
			tfStateFilename string
			tfState         string
		)

		BeforeEach(func() {
			tfState = "some-tf-state"

			tfStateFile, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = tfStateFile.Write([]byte(tfState))
			Expect(err).NotTo(HaveOccurred())

			tfStateFilename = tfStateFile.Name()
		})

		It("returns the tfState", func() {
			executorApplyError := terraform.NewExecutorApplyError(tfStateFilename, nil, true)

			actualTFState, err := executorApplyError.TFState()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualTFState).To(Equal(tfState))
		})

		Context("failure cases", func() {
			It("returns an error when tf state file does not exist", func() {
				executorApplyError := terraform.NewExecutorApplyError("/fake/file/name", nil, true)

				_, err := executorApplyError.TFState()
				Expect(err.Error()).To(ContainSubstring("no such file or directory"))
			})
		})
	})
})
