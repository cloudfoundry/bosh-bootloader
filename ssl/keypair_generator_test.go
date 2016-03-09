package ssl_test

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPairGenerator", func() {
	var (
		now   time.Time
		clock func() time.Time
	)

	BeforeEach(func() {
		now = time.Now().UTC()
		clock = func() time.Time {
			return now
		}
	})

	Describe("Generate", func() {
		It("generates an SSL certificate", func() {
			generator := ssl.NewKeyPairGenerator(clock, rsa.GenerateKey, x509.CreateCertificate)

			keyPair, err := generator.Generate("127.0.0.1")
			Expect(err).NotTo(HaveOccurred())

			tlsCert, err := tls.X509KeyPair(keyPair.Certificate, keyPair.PrivateKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(tlsCert.Certificate).To(HaveLen(1))

			parsedCerts, err := x509.ParseCertificates(tlsCert.Certificate[0])
			Expect(err).NotTo(HaveOccurred())
			Expect(parsedCerts).To(HaveLen(1))
			parsedCert := parsedCerts[0]

			err = parsedCert.VerifyHostname("127.0.0.1")
			Expect(err).NotTo(HaveOccurred())

			pkeyDER, rest := pem.Decode(keyPair.PrivateKey)
			Expect(rest).To(HaveLen(0))

			privateKey, err := x509.ParsePKCS1PrivateKey(pkeyDER.Bytes)
			Expect(err).NotTo(HaveOccurred())

			err = privateKey.Validate()
			Expect(err).NotTo(HaveOccurred())

			Expect(privateKey.Public()).To(Equal(parsedCert.PublicKey))
		})

		Context("failure cases", func() {
			It("returns an error when the rsa key cannot be generated", func() {
				fakeGenerateKey := func(random io.Reader, bits int) (priv *rsa.PrivateKey, err error) {
					return nil, errors.New("failed to generate a key")
				}

				generator := ssl.NewKeyPairGenerator(clock, fakeGenerateKey, x509.CreateCertificate)
				_, err := generator.Generate("127.0.0.1")

				Expect(err).To(MatchError("failed to generate a key"))
			})

			It("returns an error when the certificate cannot be created", func() {
				fakeCreateCertificate := func(rand io.Reader, template, parent *x509.Certificate, pub, priv interface{}) (cert []byte, err error) {
					return nil, errors.New("failed to generate a cert")
				}

				generator := ssl.NewKeyPairGenerator(clock, rsa.GenerateKey, fakeCreateCertificate)
				_, err := generator.Generate("127.0.0.1")

				Expect(err).To(MatchError("failed to generate a cert"))
			})
		})
	})
})
