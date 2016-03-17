package helpers_test

import (
	"crypto/rand"
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UUIDGenerator", func() {
	Describe("Generate", func() {
		It("generates random UUID values", func() {
			generator := helpers.NewUUIDGenerator(rand.Reader)
			uuid, err := generator.Generate()
			Expect(err).NotTo(HaveOccurred())
			Expect(uuid).To(MatchRegexp(`\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`))

			var uuids []string
			for i := 0; i < 10; i++ {
				uuid, err := generator.Generate()
				Expect(err).NotTo(HaveOccurred())
				uuids = append(uuids, uuid)
			}
			Expect(HasUniqueValues(uuids)).To(BeTrue())
		})

		Context("failure cases", func() {
			It("returns an error when the reader fails", func() {
				reader := &fakes.Reader{}
				generator := helpers.NewUUIDGenerator(reader)
				reader.ReadCall.Returns.Error = errors.New("reader failed")

				_, err := generator.Generate()
				Expect(err).To(MatchError("reader failed"))
			})
		})
	})
})

func HasUniqueValues(values []string) bool {
	valueMap := make(map[string]struct{})

	for _, value := range values {
		valueMap[value] = struct{}{}
	}

	return len(valueMap) == len(values)
}
