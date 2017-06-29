package prettyprint_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/gomegamatchers/internal/diff"
	"github.com/pivotal-cf-experimental/gomegamatchers/internal/prettyprint"
)

var _ = Describe("ExpectationFailure", func() {
	It("formats complex diffs correctly", func() {
		failure := prettyprint.ExpectationFailure(diff.MapNested{
			Key: "colors",
			NestedDifference: diff.SliceNested{
				Index: 2,
				NestedDifference: diff.PrimitiveValueMismatch{
					ExpectedValue: "blue",
					ActualValue:   "red",
				},
			},
		})

		Expect(failure).To(ContainSubstring("error at [colors][2]:"))
		Expect(failure).To(ContainSubstring("  value mismatch:"))
		Expect(failure).To(ContainSubstring("    Expected"))
		Expect(failure).To(ContainSubstring("        <string> red"))
		Expect(failure).To(ContainSubstring("    to equal"))
		Expect(failure).To(ContainSubstring("        <string> blue"))
	})

	It("returns an empty string when there is no difference", func() {
		failure := prettyprint.ExpectationFailure(diff.NoDifference{})
		Expect(failure).To(Equal("error at "))
	})

	Context("when printing primitives", func() {
		It("formats type mismatches correctly", func() {
			failure := prettyprint.ExpectationFailure(diff.PrimitiveTypeMismatch{
				ExpectedType: reflect.TypeOf(1),
				ActualValue:  "red",
			})

			Expect(failure).To(ContainSubstring("error at :"))
			Expect(failure).To(ContainSubstring("  type mismatch:"))
			Expect(failure).To(ContainSubstring("    Expected"))
			Expect(failure).To(ContainSubstring("        <string> red"))
			Expect(failure).To(ContainSubstring("    to be of type"))
			Expect(failure).To(ContainSubstring("        <int>"))
		})

		It("formats value mismatches correctly", func() {
			failure := prettyprint.ExpectationFailure(diff.PrimitiveValueMismatch{
				ExpectedValue: "blue",
				ActualValue:   "red",
			})

			Expect(failure).To(ContainSubstring("error at :"))
			Expect(failure).To(ContainSubstring("  value mismatch:"))
			Expect(failure).To(ContainSubstring("    Expected"))
			Expect(failure).To(ContainSubstring("        <string> red"))
			Expect(failure).To(ContainSubstring("    to equal"))
			Expect(failure).To(ContainSubstring("        <string> blue"))
		})
	})

	Context("when printing maps", func() {
		It("formats nested differences correctly", func() {
			failure := prettyprint.ExpectationFailure(diff.MapNested{
				Key: "age",
				NestedDifference: diff.PrimitiveValueMismatch{
					ExpectedValue: 11,
					ActualValue:   12,
				},
			})

			Expect(failure).To(ContainSubstring("error at [age]:"))
			Expect(failure).To(ContainSubstring("  value mismatch:"))
			Expect(failure).To(ContainSubstring("    Expected"))
			Expect(failure).To(ContainSubstring("        <int> 12"))
			Expect(failure).To(ContainSubstring("    to equal"))
			Expect(failure).To(ContainSubstring("        <int> 11"))
		})

		It("formats extra key differences correctly", func() {
			failure := prettyprint.ExpectationFailure(diff.MapExtraKey{
				ExtraKey: 1,
				AllKeys: []reflect.Value{
					reflect.ValueOf(1),
					reflect.ValueOf("blue"),
				},
			})

			Expect(failure).To(ContainSubstring("error at :"))
			Expect(failure).To(ContainSubstring("  extra key found:"))
			Expect(failure).To(ContainSubstring("    Expected"))
			Expect(failure).To(ContainSubstring("        [<int> 1, <string> blue]"))
			Expect(failure).To(ContainSubstring("    not to contain key"))
			Expect(failure).To(ContainSubstring("        <int> 1"))
		})

		It("formats missing key differences correctly", func() {
			failure := prettyprint.ExpectationFailure(diff.MapMissingKey{
				MissingKey: true,
				AllKeys: []reflect.Value{
					reflect.ValueOf("red"),
					reflect.ValueOf("blue"),
				},
			})

			Expect(failure).To(ContainSubstring("error at :"))
			Expect(failure).To(ContainSubstring("  missing key:"))
			Expect(failure).To(ContainSubstring("    Expected"))
			Expect(failure).To(ContainSubstring("        [<string> red, <string> blue]"))
			Expect(failure).To(ContainSubstring("    to contain key"))
			Expect(failure).To(ContainSubstring("        <bool> true"))
		})
	})

	Context("when printing slices", func() {
		It("formats nested differences correctly", func() {
			failure := prettyprint.ExpectationFailure(diff.SliceNested{
				Index: 4,
				NestedDifference: diff.PrimitiveTypeMismatch{
					ExpectedType: reflect.TypeOf(1),
					ActualValue:  "red",
				},
			})

			Expect(failure).To(ContainSubstring("error at [4]:"))
			Expect(failure).To(ContainSubstring("  type mismatch:"))
			Expect(failure).To(ContainSubstring("    Expected"))
			Expect(failure).To(ContainSubstring("        <string> red"))
			Expect(failure).To(ContainSubstring("    to be of type"))
			Expect(failure).To(ContainSubstring("        <int>"))
		})

		It("formats extra element differences correctly", func() {
			failure := prettyprint.ExpectationFailure(diff.SliceExtraElements{
				ExtraElements: reflect.ValueOf([]int{3, 4}),
				AllElements:   reflect.ValueOf([]int{1, 2, 3, 4}),
			})

			Expect(failure).To(ContainSubstring("error at :"))
			Expect(failure).To(ContainSubstring("  extra elements found:"))
			Expect(failure).To(ContainSubstring("    Expected"))
			Expect(failure).To(ContainSubstring("        [<int> 1, <int> 2, <int> 3, <int> 4]"))
			Expect(failure).To(ContainSubstring("    not to contain elements"))
			Expect(failure).To(ContainSubstring("        [<int> 3, <int> 4]"))
		})

		It("formats missing element differences correctly", func() {
			failure := prettyprint.ExpectationFailure(diff.SliceMissingElements{
				MissingElements: reflect.ValueOf([]int{3, 4}),
				AllElements:     reflect.ValueOf([]int{1, 2}),
			})

			Expect(failure).To(ContainSubstring("error at :"))
			Expect(failure).To(ContainSubstring("  missing elements:"))
			Expect(failure).To(ContainSubstring("    Expected"))
			Expect(failure).To(ContainSubstring("        [<int> 1, <int> 2]"))
			Expect(failure).To(ContainSubstring("    to contain elements"))
			Expect(failure).To(ContainSubstring("        [<int> 3, <int> 4]"))
		})
	})
})
