package ssl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
)

var _ = Describe("KeyPair", func() {
	Describe("IsEmpty", func() {
		It("returns true if the keypair is empty", func() {
			keyPair := ssl.KeyPair{}

			Expect(keyPair.IsEmpty()).To(BeTrue())
		})

		It("returns false if the keypair is not empty", func() {
			keyPair := ssl.KeyPair{
				Certificate: []byte("some-cert"),
				PrivateKey:  []byte("some-key"),
			}

			Expect(keyPair.IsEmpty()).To(BeFalse())
		})
	})
})
