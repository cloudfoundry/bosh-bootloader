package storage_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-bootloader/storage"
)

var _ = Describe("Jumpbox", func() {
	Describe("IsEmpty", func() {
		It("returns true when the struct is empty", func() {
			Expect(Jumpbox{}.IsEmpty()).To(BeTrue())
		})
		It("returns false when the struct is populated", func() {
			Expect(Jumpbox{URL: ":)"}.IsEmpty()).To(BeFalse())
		})
	})
	Describe("GetURLWithJumpboxUser", func() {
		It("returns a jumpbox URL with default user when no user is specified in the URL", func() {
			jumpbox := Jumpbox{URL: "1.1.1.1:22"}
			Expect(jumpbox.GetURLWithJumpboxUser()).To(Equal("jumpbox@1.1.1.1:22"))
		})
		It("returns a jumpbox URL as is when the user is specified in the URL", func() {
			jumpbox := Jumpbox{URL: "ubuntu@1.1.1.1:22"}
			Expect(jumpbox.GetURLWithJumpboxUser()).To(Equal("ubuntu@1.1.1.1:22"))
		})
	})
})
