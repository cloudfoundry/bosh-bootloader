package gomegamatchers_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/gomegamatchers"
)

type animalStringer struct {
	Data string
}

func (a animalStringer) String() string {
	return a.Data
}

var _ = Describe("MatchYAMLMatcher", func() {
	var animals, plants string

	BeforeEach(func() {
		animals = "- cats:\n  - lion\n- fish:\n  - salmon"
		plants = "- tropical:\n  - palm\n- desert:\n  - cactus"
	})

	Describe("Match", func() {
		It("works with complex yaml", func() {
			yaml, err := ioutil.ReadFile("fixtures/santa_monica_correct.yml")
			Expect(err).NotTo(HaveOccurred())

			isMatch, err := gomegamatchers.MatchYAML(yaml).Match(yaml)
			Expect(err).NotTo(HaveOccurred())
			Expect(isMatch).To(BeTrue())
		})

		Context("when arguments are strings", func() {
			It("returns true when the YAML matches", func() {
				isMatch, err := gomegamatchers.MatchYAML(animals).Match(animals)
				Expect(err).NotTo(HaveOccurred())
				Expect(isMatch).To(BeTrue())
			})

			It("returns false when the YAML does not match", func() {
				isMatch, err := gomegamatchers.MatchYAML(animals).Match(plants)
				Expect(err).NotTo(HaveOccurred())
				Expect(isMatch).To(BeFalse())
			})
		})

		Context("when an input is a byte slice", func() {
			var animalBytes []byte

			BeforeEach(func() {
				animalBytes = []byte(animals)
			})

			It("returns true when the YAML matches", func() {
				isMatch, err := gomegamatchers.MatchYAML(animalBytes).Match(animals)
				Expect(err).NotTo(HaveOccurred())
				Expect(isMatch).To(BeTrue())
			})

			It("returns false when the YAML does not match", func() {
				isMatch, err := gomegamatchers.MatchYAML(animalBytes).Match(plants)
				Expect(err).NotTo(HaveOccurred())
				Expect(isMatch).To(BeFalse())
			})
		})

		Context("when an input is a Stringer", func() {
			var stringer animalStringer

			BeforeEach(func() {
				stringer = animalStringer{
					Data: animals,
				}
			})

			It("returns true when the YAML matches", func() {
				isMatch, err := gomegamatchers.MatchYAML(stringer).Match(animals)
				Expect(err).NotTo(HaveOccurred())
				Expect(isMatch).To(BeTrue())
			})

			It("returns false when the YAML does not match", func() {
				isMatch, err := gomegamatchers.MatchYAML(stringer).Match(plants)
				Expect(err).NotTo(HaveOccurred())
				Expect(isMatch).To(BeFalse())
			})
		})

		Describe("errors", func() {
			It("returns an error when one of the inputs is not a string, byte slice, or Stringer", func() {
				_, err := gomegamatchers.MatchYAML(animals).Match(123213)
				Expect(err.Error()).To(ContainSubstring("MatchYAMLMatcher matcher requires a string or stringer."))
				Expect(err.Error()).To(ContainSubstring("Got:\n    <int>: 123213"))
			})

			It("returns an error when the YAML is invalid", func() {
				_, err := gomegamatchers.MatchYAML(animals).Match("some: invalid: yaml")
				Expect(err.Error()).To(ContainSubstring("mapping values are not allowed in this context"))
			})
		})
	})

	Describe("FailureMessage", func() {
		It("returns a failure message", func() {
			message := gomegamatchers.MatchYAML("a: 1").FailureMessage("b: 2")
			Expect(message).To(ContainSubstring("error at :"))
			Expect(message).To(ContainSubstring("  extra key found:"))
			Expect(message).To(ContainSubstring("    Expected"))
			Expect(message).To(ContainSubstring("        [<string> b]"))
			Expect(message).To(ContainSubstring("    not to contain key"))
			Expect(message).To(ContainSubstring("        <string> b"))
		})

		It("provides localized error information", func() {
			correctYAML, err := ioutil.ReadFile("fixtures/santa_monica_correct.yml")
			Expect(err).NotTo(HaveOccurred())

			incorrectYAML, err := ioutil.ReadFile("fixtures/santa_monica_incorrect.yml")
			Expect(err).NotTo(HaveOccurred())

			message := gomegamatchers.MatchYAML(string(correctYAML)).FailureMessage(string(incorrectYAML))
			Expect(message).To(
				SatisfyAny(
					SatisfyAll(
						ContainSubstring("error at [population][1980][absolute]:"),
						ContainSubstring("  value mismatch:"),
						ContainSubstring("    Expected"),
						ContainSubstring("        <int> 999999999"),
						ContainSubstring("    to equal"),
						ContainSubstring("        <int> 88314"),
					),
					SatisfyAll(
						ContainSubstring("error at [population][2000][absolute]:"),
						ContainSubstring("  type mismatch:"),
						ContainSubstring("    Expected"),
						ContainSubstring("        <string> wrong type"),
						ContainSubstring("    to be of type"),
						ContainSubstring("        <int>"),
					),
					SatisfyAll(
						ContainSubstring("error at [population][1990]:"),
						ContainSubstring("  extra key found:"),
						ContainSubstring("    Expected"),
						ContainSubstring("        [<string> growth_rate, <string> wrong_key]"),
						ContainSubstring("    not to contain key"),
						ContainSubstring("        <string> wrong_key"),
					),
					SatisfyAll(
						ContainSubstring("error at [population][1990]:"),
						ContainSubstring("  extra key found:"),
						ContainSubstring("    Expected"),
						ContainSubstring("        [<string> wrong_key, <string> growth_rate]"),
						ContainSubstring("    not to contain key"),
						ContainSubstring("        <string> wrong_key"),
					),
				))
		})

		Describe("errors", func() {
			It("returns the error as the message", func() {
				message := gomegamatchers.MatchYAML(animals).FailureMessage("some: invalid: yaml")
				Expect(message).To(ContainSubstring("mapping values are not allowed in this context"))
			})
		})
	})

	Describe("NegatedFailureMessage", func() {
		It("returns a negated failure message", func() {
			message := gomegamatchers.MatchYAML("a: 1").NegatedFailureMessage("b: 2")
			Expect(message).To(ContainSubstring("Expected"))
			Expect(message).To(ContainSubstring("<string>: b: 2"))
			Expect(message).To(ContainSubstring("not to match YAML of"))
			Expect(message).To(ContainSubstring("<string>: a: 1"))
		})
	})
})
