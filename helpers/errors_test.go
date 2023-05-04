package helpers_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errors", func() {
	Describe("Error", func() {
		Context("when there is one error", func() {
			It("returns the error string", func() {
				errors := helpers.NewErrors("this is an error")
				Expect(errors.Error()).To(Equal("this is an error"))
			})
		})

		Context("when there are multiple errors", func() {
			It("returns a string representation of the errors ", func() {
				errors := helpers.NewErrors("this is an error", "this is also an error")
				Expect(errors.Error()).To(Equal("the following errors occurred:\nthis is an error,\nthis is also an error"))
			})
		})
	})

	Describe("Add", func() {
		It("adds a errors to the list", func() {
			errList := helpers.NewErrors()
			errList.Add(errors.New("foo"))
			errList.Add(errors.New("bar"))

			Expect(errList.Error()).To(Equal("the following errors occurred:\nfoo,\nbar"))
		})
	})
})
