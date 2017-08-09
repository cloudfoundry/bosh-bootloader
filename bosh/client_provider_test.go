package bosh_test

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client Provider", func() {
	Describe("Client", func() {
		var (
			clientProvider bosh.ClientProvider
		)

		BeforeEach(func() {
			clientProvider = bosh.NewClientProvider()
		})

		It("returns a bosh client", func() {
			boshClient := clientProvider.Client(false, "some-director-address", "some-director-username", "some-director-password", "some-fake-ca")

			_, ok := boshClient.(bosh.Client)
			Expect(ok).To(BeTrue())
		})
	})
})
