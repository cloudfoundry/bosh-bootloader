package ec2_test

import (
	"crypto/rand"
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/ec2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type FakeReader struct {
	ReadCall struct {
		Returns struct {
			Error error
		}
	}
}

func (r *FakeReader) Read([]byte) (int, error) {
	return 0, r.ReadCall.Returns.Error
}

var _ = Describe("UUIDGenerator", func() {
	Describe("Generate", func() {
		It("generates random UUID values", func() {
			generator := ec2.NewUUIDGenerator(rand.Reader)
			uuid, err := generator.Generate()
			Expect(err).NotTo(HaveOccurred())
			Expect(uuid).To(MatchRegexp(`\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`))

			var uuids []string
			for i := 0; i < 10; i++ {
				uuid, err := generator.Generate()
				Expect(err).NotTo(HaveOccurred())
				uuids = append(uuids, uuid)
			}
			Expect(HasUniqueUUIDs(uuids)).To(BeTrue())
		})

		Context("failure cases", func() {
			It("returns an error when the reader fails", func() {
				reader := &FakeReader{}
				generator := ec2.NewUUIDGenerator(reader)
				reader.ReadCall.Returns.Error = errors.New("reader failed")

				_, err := generator.Generate()
				Expect(err).To(MatchError("reader failed"))
			})
		})
	})
})

func HasUniqueUUIDs(uuids []string) bool {
	values := make(map[string]struct{})

	for _, uuid := range uuids {
		values[uuid] = struct{}{}
	}

	return len(values) == len(uuids)
}
