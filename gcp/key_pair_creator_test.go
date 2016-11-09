package gcp_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"

	"github.com/cloudfoundry/bosh-bootloader/gcp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
)

var _ = Describe("KeyPairCreator", func() {
	var (
		keyPairCreator gcp.KeyPairCreator
	)

	It("generates a keypair", func() {
		keyPairCreator = gcp.NewKeyPairCreator(rand.Reader, rsa.GenerateKey, ssh.NewPublicKey)

		privateKey, publicKey, err := keyPairCreator.Create()
		Expect(err).NotTo(HaveOccurred())
		Expect(privateKey).NotTo(BeEmpty())
		Expect(publicKey).NotTo(BeEmpty())

		pemBlock, rest := pem.Decode([]byte(privateKey))
		Expect(rest).To(HaveLen(0))
		Expect(pemBlock.Type).To(Equal("RSA PRIVATE KEY"))

		parsedPrivateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		Expect(err).NotTo(HaveOccurred())

		err = parsedPrivateKey.Validate()
		Expect(err).NotTo(HaveOccurred())

		newPublicKey, err := ssh.NewPublicKey(parsedPrivateKey.Public())
		Expect(err).NotTo(HaveOccurred())

		Expect(string(ssh.MarshalAuthorizedKey(newPublicKey))).To(Equal(publicKey))
	})

	Context("failure cases", func() {
		It("returns an error when the rsaKeyGenerator fails", func() {
			keyPairCreator = gcp.NewKeyPairCreator(rand.Reader,
				func(_ io.Reader, _ int) (*rsa.PrivateKey, error) {
					return nil, errors.New("rsa key generator failed")
				},
				ssh.NewPublicKey)

			_, _, err := keyPairCreator.Create()
			Expect(err).To(MatchError("rsa key generator failed"))
		})

		It("returns an error when the ssh public key generator fails", func() {
			keyPairCreator = gcp.NewKeyPairCreator(rand.Reader, rsa.GenerateKey,
				func(_ interface{}) (ssh.PublicKey, error) {
					return nil, errors.New("ssh public key gen failed")
				})

			_, _, err := keyPairCreator.Create()
			Expect(err).To(MatchError("ssh public key gen failed"))
		})
	})
})
