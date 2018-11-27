package application_test

import (
	"github.com/cloudfoundry/bosh-bootloader/application"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StringSlice", func() {
	Describe("ContainsAll", func() {
		It("returns true if the slice contains all the targets", func() {
			stringSlice := application.StringSlice{"apple", "banana", "cat", "dog", "elephant"}
			Expect(stringSlice.ContainsAny("apple", "zebra")).To(BeTrue())
		})

		It("return false if the slice does not contain any the target", func() {
			stringSlice := application.StringSlice{"apple", "banana", "cat", "dog", "elephant"}
			Expect(stringSlice.ContainsAny("zebra", "kangaroo")).To(BeFalse())
		})
	})
})
