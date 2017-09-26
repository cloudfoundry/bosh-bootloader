package terraform_test

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExecutorError", func() {
	Describe("Error", func() {
		Context("when debug is true", func() {
			It("returns just the internal error message", func() {
				err := errors.New("some-error")
				executorError := terraform.NewExecutorError("", err, true)

				Expect(executorError.Error()).To(Equal(err.Error()))
			})
		})

		Context("when debug is false", func() {
			It("returns the internal error message and mentions the --debug flag", func() {
				err := errors.New("some-error")
				executorError := terraform.NewExecutorError("", err, false)

				Expect(executorError.Error()).To(Equal(fmt.Sprintf("%s\n%s", err.Error(), "Some output has been redacted, use `bbl latest-error` to see it or run again with --debug for additional debug output")))
			})
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
			executorError := terraform.NewExecutorError(tfStateFilename, nil, true)

			actualTFState, err := executorError.TFState()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualTFState).To(Equal(tfState))
		})

		Context("failure cases", func() {
			Context("when tf state file does not exist", func() {
				It("returns an error", func() {
					executorError := terraform.NewExecutorError("/fake/file/name", nil, true)
					_, err := executorError.TFState()
					Expect(err.Error()).To(ContainSubstring("no such file or directory"))
				})
			})
		})
	})
})
