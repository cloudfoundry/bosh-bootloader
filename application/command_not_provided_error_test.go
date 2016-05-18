package application_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
)

var _ = Describe("CommandNotProvidedError", func() {
	Describe("Error", func() {
		It("returns the formatted raw error", func() {
			commandNotProvidedError := application.NewCommandNotProvidedError()
			Expect(commandNotProvidedError.Error()).To(Equal("unknown command: [EMPTY]"))
		})
	})
})
