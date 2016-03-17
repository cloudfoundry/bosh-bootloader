package helpers_test

import (
	"crypto/rand"
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StringGenerator", func() {
	Describe("Generate", func() {
		It("generates random alphanumeric values of a given length", func() {
			generator := helpers.NewStringGenerator(rand.Reader)
			randomString, err := generator.Generate("prefix-", 15)
			Expect(err).NotTo(HaveOccurred())
			Expect(randomString).To(MatchRegexp(`prefix-\w{15}`))

			var randomStrings []string
			for i := 0; i < 10; i++ {
				randomString, err := generator.Generate("prefix-", 15)
				Expect(err).NotTo(HaveOccurred())
				randomStrings = append(randomStrings, randomString)
			}
			Expect(HasUniqueValues(randomStrings)).To(BeTrue())
		})

		Context("failure cases", func() {
			It("returns an error when the reader fails", func() {
				reader := &fakes.Reader{}
				generator := helpers.NewStringGenerator(reader)
				reader.ReadCall.Returns.Error = errors.New("reader failed")

				_, err := generator.Generate("prefix", 1)
				Expect(err).To(MatchError("reader failed"))
			})
		})
	})
})
