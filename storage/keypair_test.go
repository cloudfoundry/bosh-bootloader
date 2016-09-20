package storage_test

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPair", func() {
	Describe("IsEmpty", func() {
		It("returns true if empty", func() {
			keypair := storage.KeyPair{}

			Expect(keypair.IsEmpty()).To(BeTrue())
		})

		It("returns false if not empty", func() {
			keypair := storage.KeyPair{
				Name: "some-name",
			}

			Expect(keypair.IsEmpty()).To(BeFalse())
		})
	})
})
