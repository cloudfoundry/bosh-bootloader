package deepequal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/gomegamatchers/internal/deepequal"
	"github.com/pivotal-cf-experimental/gomegamatchers/internal/diff"
)

var _ = Describe("Primitive", func() {
	It("returns true when comparing equal primitives", func() {
		equal, difference := deepequal.Primitive(1, 1)

		Expect(equal).To(BeTrue())
		Expect(difference).To(Equal(diff.NoDifference{}))
	})

	It("returns a diff when comparing unequal primitives", func() {
		equal, difference := deepequal.Primitive(true, false)

		Expect(equal).To(BeFalse())
		Expect(difference).To(Equal(diff.PrimitiveValueMismatch{
			ExpectedValue: true,
			ActualValue:   false,
		}))
	})
})
