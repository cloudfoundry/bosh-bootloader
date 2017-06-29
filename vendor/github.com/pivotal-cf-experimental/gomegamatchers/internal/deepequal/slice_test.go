package deepequal_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/gomegamatchers/internal/deepequal"
	"github.com/pivotal-cf-experimental/gomegamatchers/internal/diff"
)

var _ = Describe("Slice", func() {
	It("returns true when the lengths and values match", func() {
		slice := reflect.ValueOf([]int{1, 2, 3, 4})

		equal, difference := deepequal.Slice(slice, slice)
		Expect(equal).To(BeTrue())
		Expect(difference).To(Equal(diff.NoDifference{}))
	})

	It("returns a diff when the lengths match but the values do not", func() {
		expected := reflect.ValueOf([]int{1, 2, 3, 4})
		actual := reflect.ValueOf([]int{1, 2, 0, 4})

		equal, difference := deepequal.Slice(expected, actual)

		Expect(equal).To(BeFalse())
		Expect(difference).To(Equal(diff.SliceNested{
			Index: 2,
			NestedDifference: diff.PrimitiveValueMismatch{
				ExpectedValue: 3,
				ActualValue:   0,
			},
		}))
	})

	It("returns a diff when the actual slice contains values that are not in the expected slice", func() {
		expected := reflect.ValueOf([]int{1, 2})
		actual := reflect.ValueOf([]int{1, 2, 3, 4})

		equal, difference := deepequal.Slice(expected, actual)
		Expect(equal).To(BeFalse())

		extraElementsDifference, isSliceExtraElements := difference.(diff.SliceExtraElements)
		Expect(isSliceExtraElements).To(BeTrue())

		Expect(extraElementsDifference.ExtraElements.Len()).To(Equal(2))
		Expect(int(extraElementsDifference.ExtraElements.Index(0).Int())).To(Equal(3))
		Expect(int(extraElementsDifference.ExtraElements.Index(1).Int())).To(Equal(4))

		Expect(extraElementsDifference.AllElements.Len()).To(Equal(4))
		Expect(int(extraElementsDifference.AllElements.Index(0).Int())).To(Equal(1))
		Expect(int(extraElementsDifference.AllElements.Index(1).Int())).To(Equal(2))
		Expect(int(extraElementsDifference.AllElements.Index(2).Int())).To(Equal(3))
		Expect(int(extraElementsDifference.AllElements.Index(3).Int())).To(Equal(4))
	})

	It("returns a diff when the expected slice contains values that are not in the actual slice", func() {
		expected := reflect.ValueOf([]int{1, 2, 3, 4})
		actual := reflect.ValueOf([]int{1, 2})

		equal, difference := deepequal.Slice(expected, actual)
		Expect(equal).To(BeFalse())

		missingElementsDifference, isSliceMissingElements := difference.(diff.SliceMissingElements)
		Expect(isSliceMissingElements).To(BeTrue())

		Expect(missingElementsDifference.MissingElements.Len()).To(Equal(2))
		Expect(int(missingElementsDifference.MissingElements.Index(0).Int())).To(Equal(3))
		Expect(int(missingElementsDifference.MissingElements.Index(1).Int())).To(Equal(4))

		Expect(missingElementsDifference.AllElements.Len()).To(Equal(2))
		Expect(int(missingElementsDifference.AllElements.Index(0).Int())).To(Equal(1))
		Expect(int(missingElementsDifference.AllElements.Index(1).Int())).To(Equal(2))
	})
})
