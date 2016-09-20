package bosh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
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
			boshClient := clientProvider.Client("some-director-address", "some-director-username", "some-director-password")

			_, ok := boshClient.(bosh.Client)
			Expect(ok).To(BeTrue())
		})
	})
})
