package application_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
)

var _ = Describe("InvalidFlagError", func() {
	Describe("Error", func() {
		It("returns the formatted raw error", func() {
			invalidFlagError := application.NewInvalidFlagError(errors.New("some-invalid-flag-error"))
			Expect(invalidFlagError.Error()).To(Equal("some-invalid-flag-error"))
		})
	})
})
