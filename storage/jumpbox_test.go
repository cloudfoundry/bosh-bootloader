package storage_test

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JumpboxDeployment", func() {
	Describe("IsEmpty", func() {
		It("returns true if empty", func() {
			jumpbox := storage.JumpboxDeployment{}

			Expect(jumpbox.IsEmpty()).To(BeTrue())
		})

		It("returns false if not empty", func() {
			jumpbox := storage.JumpboxDeployment{
				Variables: "some-vars",
				Manifest:  "some-name",
				State: map[string]interface{}{
					"key":  1,
					"key2": "value",
				},
			}

			Expect(jumpbox.IsEmpty()).To(BeFalse())
		})
	})
})
