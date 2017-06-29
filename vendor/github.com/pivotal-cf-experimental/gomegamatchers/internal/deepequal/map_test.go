package deepequal_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/gomegamatchers/internal/deepequal"
	"github.com/pivotal-cf-experimental/gomegamatchers/internal/diff"
)

var _ = Describe("Map", func() {
	It("returns no difference when the keys and values match (regardless of the order of the keys)", func() {
		expected := reflect.ValueOf(map[string]int{"a": 1, "b": 2, "c": 3})
		actual := reflect.ValueOf(map[string]int{"c": 3, "b": 2, "a": 1})

		equal, difference := deepequal.Map(expected, actual)
		Expect(equal).To(BeTrue())
		Expect(difference).To(Equal(diff.NoDifference{}))
	})

	It("returns a diff when keys match but values do not", func() {
		expected := reflect.ValueOf(map[string]int{"a": 1, "b": 2, "c": 3})
		actual := reflect.ValueOf(map[string]int{"a": 1, "b": 0, "c": 3})

		equal, difference := deepequal.Map(expected, actual)
		Expect(equal).To(BeFalse())
		Expect(difference).To(Equal(diff.MapNested{
			Key: "b",
			NestedDifference: diff.PrimitiveValueMismatch{
				ExpectedValue: 2,
				ActualValue:   0,
			},
		}))
	})

	It("returns a diff when the actual map contains keys that are not in the expected map", func() {
		expected := reflect.ValueOf(map[string]int{"a": 1})
		actual := reflect.ValueOf(map[string]int{"a": 1, "b": 2})

		equal, difference := deepequal.Map(expected, actual)
		Expect(equal).To(BeFalse())

		extraKeyDifference, isMapExtraKey := difference.(diff.MapExtraKey)
		Expect(isMapExtraKey).To(BeTrue())
		Expect(extraKeyDifference.ExtraKey).To(Equal("b"))

		Expect([]string{
			extraKeyDifference.AllKeys[0].String(),
			extraKeyDifference.AllKeys[1].String(),
		}).To(ConsistOf([]string{"a", "b"}))
	})

	It("returns a diff when the expected map contains keys that are not in the actual map", func() {
		expected := reflect.ValueOf(map[string]int{"a": 1, "b": 2})
		actual := reflect.ValueOf(map[string]int{"a": 1})

		equal, difference := deepequal.Map(expected, actual)
		Expect(equal).To(BeFalse())

		missingKeyDifference, isMapMissingKey := difference.(diff.MapMissingKey)
		Expect(isMapMissingKey).To(BeTrue())

		Expect(missingKeyDifference.MissingKey).To(Equal("b"))
		Expect(missingKeyDifference.AllKeys).To(HaveLen(1))
		Expect(missingKeyDifference.AllKeys[0].String()).To(Equal("a"))
	})
})
