package ec2_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPairVerifier", func() {
	Describe("Verify", func() {
		var verifier ec2.KeyPairVerifier

		It("returns with no error if the fingerprints match", func() {
			err := verifier.Verify(fingerprint, []byte(validRSAKey))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			It("returns an error when the fingerprints do not match", func() {
				invalidFingerprint := "some-fingerprint"
				err := verifier.Verify(invalidFingerprint, []byte(validRSAKey))
				Expect(err).To(MatchError("the local keypair fingerprint does not match the keypair fingerprint on AWS"))
			})

			Context("when the keypair is not PEM encoded", func() {
				It("returns an error", func() {
					err := verifier.Verify(fingerprint, []byte("not-valid-pem-data"))
					Expect(err).To(MatchError("the local keypair does not contain a valid PEM encoded private key, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance."))
				})
			})

			Context("when the keypair does not contain a valid RSA key", func() {
				It("returns an error", func() {
					err := verifier.Verify(fingerprint, []byte(invalidRSAKey))
					Expect(err).To(MatchError("the local keypair does not contain a valid rsa private key, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance."))
				})
			})
		})
	})
})
