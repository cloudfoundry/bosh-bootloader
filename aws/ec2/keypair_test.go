package ec2_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
)

var _ = Describe("KeyPair", func() {
	Describe("IsEmpty", func() {
		It("returns true if the keypair is empty", func() {
			keypair := ec2.KeyPair{}

			Expect(keypair.IsEmpty()).To(BeTrue())
		})

		It("returns false if the keypair is not empty", func() {
			keypair := ec2.KeyPair{
				Name:       "key-name",
				PublicKey:  "public",
				PrivateKey: "private",
			}

			Expect(keypair.IsEmpty()).To(BeFalse())
		})
	})
})
