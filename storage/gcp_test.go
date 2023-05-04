package storage_test

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCP", func() {
	Describe("Empty", func() {
		var gcp storage.GCP
		Context("when all fields are blank", func() {
			BeforeEach(func() {
				gcp = storage.GCP{}
			})

			It("returns true", func() {
				empty := gcp.Empty()
				Expect(empty).To(BeTrue())
			})
		})

		Context("when at least one field is present", func() {
			BeforeEach(func() {
				gcp = storage.GCP{ServiceAccountKey: "some-account-key"}
			})

			It("returns false", func() {
				empty := gcp.Empty()
				Expect(empty).To(BeFalse())
			})
		})
	})
})
