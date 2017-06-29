package gomegamatchers_test

import (
	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContainSequence", func() {
	Context("when actual is not a slice", func() {
		It("should error", func() {
			_, err := gomegamatchers.ContainSequence(func() {}).Match("not a slice")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when actual is a slice", func() {
		var sequence []string

		BeforeEach(func() {
			sequence = []string{
				"value-1",
				"value-2",
				"value-3",
			}
		})

		Context("when the entire sequence is present", func() {
			It("should succeed", func() {
				Expect([]string{
					"value-0",
					"value-1",
					"value-2",
					"value-3",
					"value-4",
				}).To(gomegamatchers.ContainSequence(sequence))
			})
		})

		Context("when some of the sequence is present", func() {
			It("should fail", func() {
				Expect([]string{
					"value-0",
					"value-1",
					"value-3",
					"value-4",
				}).NotTo(gomegamatchers.ContainSequence(sequence))
			})
		})

		Context("when none of the sequence is present", func() {
			It("should fail", func() {
				Expect([]string{
					"value-0",
					"value-4",
				}).NotTo(gomegamatchers.ContainSequence(sequence))
			})
		})

		Context("when the elements match, but the order does not", func() {
			It("should fail", func() {
				Expect([]string{
					"value-0",
					"value-3",
					"value-1",
					"value-2",
					"value-4",
				}).NotTo(gomegamatchers.ContainSequence(sequence))
			})
		})

		Context("when the sequence shows up at the end of the actual slice", func() {
			It("should succeed", func() {
				Expect([]string{
					"value-0",
					"value-1",
					"value-2",
					"value-3",
				}).To(gomegamatchers.ContainSequence(sequence))
			})
		})
	})

	Describe("FailureMessage", func() {
		It("returns an understandable error message", func() {
			Expect(gomegamatchers.ContainSequence([]int{1, 2, 3}).FailureMessage([]int{5, 6})).To(Equal("Expected\n    <[]int | len:2, cap:2>: [5, 6]\nto contain sequence\n    <[]int | len:3, cap:3>: [1, 2, 3]"))
		})
	})

	Describe("NegatedFailureMessage", func() {
		It("returns an understandable error message", func() {
			Expect(gomegamatchers.ContainSequence([]int{1, 2, 3}).NegatedFailureMessage([]int{1, 2, 3})).To(Equal("Expected\n    <[]int | len:3, cap:3>: [1, 2, 3]\nnot to contain sequence\n    <[]int | len:3, cap:3>: [1, 2, 3]"))
		})
	})
})
