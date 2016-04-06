package storage_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BOSH", func() {
	Describe("IsEmpty", func() {
		It("returns true if empty", func() {
			bosh := storage.BOSH{}

			Expect(bosh.IsEmpty()).To(BeTrue())
		})

		It("returns false if not empty", func() {
			bosh := storage.BOSH{
				DirectorUsername: "some-name",
				State: map[string]interface{}{
					"key":  1,
					"key2": "value",
				},
			}

			Expect(bosh.IsEmpty()).To(BeFalse())
		})
	})
})
