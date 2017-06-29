package deepequal_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/gomegamatchers/internal/deepequal"
	"github.com/pivotal-cf-experimental/gomegamatchers/internal/diff"
)

var _ = Describe("Compare", func() {
	types := map[string][]interface{}{
		"bool":       {true, false},
		"string":     {"a", "b"},
		"int":        {int(1), int(2)},
		"int8":       {int8(1), int8(2)},
		"int16":      {int16(1), int16(2)},
		"int32":      {int32(1), int32(2)},
		"int64":      {int64(1), int64(2)},
		"uint":       {uint(1), uint(2)},
		"uint8":      {uint8(1), uint8(2)},
		"uint16":     {uint16(1), uint16(2)},
		"uint32":     {uint32(1), uint32(2)},
		"uint64":     {uint64(1), uint64(2)},
		"uintptr":    {uintptr(1), uintptr(2)},
		"float32":    {float32(1.0), float32(2.0)},
		"float64":    {float64(1.0), float64(2.0)},
		"complex64":  {complex64(1i), complex64(2i)},
		"complex128": {complex128(1i), complex128(2i)},
	}

	It("returns true when the objects match", func() {
		someObject := map[string]interface{}{
			"a": 1,
			"b": []int{1, 2, 3, 4},
			"c": 3,
		}

		equal, difference := deepequal.Compare(someObject, someObject)
		Expect(equal).To(BeTrue())
		Expect(difference).To(Equal(diff.NoDifference{}))
	})

	It("returns a diff when the types are mismatched", func() {
		for expectedName, expectedValues := range types {
			for actualName, actualValues := range types {
				if expectedName != actualName {
					equal, difference := deepequal.Compare(expectedValues[0], actualValues[0])

					Expect(equal).To(BeFalse())
					Expect(difference).To(Equal(diff.PrimitiveTypeMismatch{
						ExpectedType: reflect.TypeOf(expectedValues[0]),
						ActualValue:  actualValues[0],
					}))
				}
			}
		}
	})

	It("returns a diff when the values are mismatched", func() {
		for _, values := range types {
			equal, difference := deepequal.Compare(values[0], values[1])

			Expect(equal).To(BeFalse())
			Expect(difference).To(Equal(diff.PrimitiveValueMismatch{
				ExpectedValue: values[0],
				ActualValue:   values[1],
			}))
		}
	})

	It("returns a nested diff when comparing nested objects", func() {
		expected := map[string]interface{}{
			"a": 1,
			"b": []int{1, 2, 3, 4},
			"c": 3,
		}
		actual := map[string]interface{}{
			"a": 1,
			"b": []int{1, 2, 0, 4},
			"c": 3,
		}

		equal, difference := deepequal.Compare(expected, actual)

		Expect(equal).To(BeFalse())
		Expect(difference).To(Equal(diff.MapNested{
			Key: "b",
			NestedDifference: diff.SliceNested{
				Index: 2,
				NestedDifference: diff.PrimitiveValueMismatch{
					ExpectedValue: 3,
					ActualValue:   0,
				},
			},
		}))
	})

	Context("when both are nil", func() {
		It("returns true", func() {
			var expected interface{}
			var actual interface{}

			equal, _ := deepequal.Compare(expected, actual)
			Expect(equal).To(BeTrue())
		})
	})

	Context("when actual is nil", func() {
		It("does not panic", func() {
			expected := map[string]interface{}{}
			var actual interface{}

			equal, difference := deepequal.Compare(expected, actual)

			Expect(equal).To(BeFalse())
			Expect(difference).To(Equal(diff.PrimitiveTypeMismatch{
				ExpectedType: reflect.TypeOf(expected),
				ActualValue:  nil,
			}))
		})
	})

	Context("when expected is nil", func() {
		It("does not panic", func() {
			var expected interface{}
			actual := map[string]interface{}{}

			equal, difference := deepequal.Compare(expected, actual)

			Expect(equal).To(BeFalse())
			Expect(difference).To(BeNil())
		})
	})
})
