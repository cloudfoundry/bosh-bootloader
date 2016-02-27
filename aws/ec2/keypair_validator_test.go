package ec2_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPairValidator", func() {
	var validator ec2.KeyPairValidator

	Describe("Validate", func() {
		It("validates a valid keypair", func() {
			err := validator.Validate([]byte(validRSAKey))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			Context("when the keypair is not PEM encoded", func() {
				It("returns an error", func() {
					err := validator.Validate([]byte("not-valid-pem-data"))
					Expect(err).To(MatchError("the local keypair does not contain a valid PEM encoded private key, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance."))
				})
			})

			Context("when the keypair does not contain a valid RSA key", func() {
				It("returns an error", func() {
					err := validator.Validate([]byte(invalidRSAKey))
					Expect(err).To(MatchError("the local keypair does not contain a valid rsa private key, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance."))
				})
			})
		})
	})
})
