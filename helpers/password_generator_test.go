package helpers_test

import (
	"crypto/rand"
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PasswordGenerator", func() {
	Describe("Generate", func() {
		It("generates random 15 alphanumeric values", func() {
			generator := helpers.NewPasswordGenerator(rand.Reader)
			password, err := generator.Generate()
			Expect(err).NotTo(HaveOccurred())
			Expect(password).To(MatchRegexp(`\w{15}`))

			var passwords []string
			for i := 0; i < 10; i++ {
				password, err := generator.Generate()
				Expect(err).NotTo(HaveOccurred())
				passwords = append(passwords, password)
			}
			Expect(HasUniqueUUIDs(passwords)).To(BeTrue())
		})

		Context("failure cases", func() {
			It("returns an error when the reader fails", func() {
				reader := &fakes.Reader{}
				generator := helpers.NewPasswordGenerator(reader)
				reader.ReadCall.Returns.Error = errors.New("reader failed")

				_, err := generator.Generate()
				Expect(err).To(MatchError("reader failed"))
			})
		})
	})
})
