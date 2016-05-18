package application_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
)

var _ = Describe("InvalidCommandError", func() {
	Describe("Error", func() {
		It("returns the formatted raw error", func() {
			invalidCommandError := application.NewInvalidCommandError(errors.New("some error"))
			Expect(invalidCommandError.Error()).To(Equal("some error"))
		})
	})
})
