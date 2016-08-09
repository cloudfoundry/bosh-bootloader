package ssl_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"

	certstrappkix "github.com/square/certstrap/pkix"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPairGenerator", func() {
	var (
		generator ssl.KeyPairGenerator

		fakePrivateKeyGenerator *fakes.PrivateKeyGenerator
		fakeCertstrapPKIX       *fakes.CertstrapPKIX

		caPrivateKey *rsa.PrivateKey
		caPublicKey  *rsa.PublicKey

		privateKey *rsa.PrivateKey
		publicKey  *rsa.PublicKey

		ca         *certstrappkix.Certificate
		csr        *certstrappkix.CertificateSigningRequest
		signedCert *certstrappkix.Certificate
		key        *certstrappkix.Key

		exportedCA   []byte
		exportedCert []byte
	)

	BeforeEach(func() {
		fakePrivateKeyGenerator = &fakes.PrivateKeyGenerator{}
		fakeCertstrapPKIX = &fakes.CertstrapPKIX{}

		generator = ssl.NewKeyPairGenerator(
			fakePrivateKeyGenerator.GenerateKey,
			fakeCertstrapPKIX.CreateCertificateAuthority,
			fakeCertstrapPKIX.CreateCertificateSigningRequest,
			fakeCertstrapPKIX.CreateCertificateHost,
		)

		var err error
		keyBlock, _ := pem.Decode([]byte(caPrivateKeyPEM))
		caPrivateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		Expect(err).NotTo(HaveOccurred())

		caPublicKey = &caPrivateKey.PublicKey
		caKey := &certstrappkix.Key{
			Private: caPrivateKey,
			Public:  caPublicKey,
		}

		keyBlock, _ = pem.Decode([]byte(privateKeyPEM))
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		Expect(err).NotTo(HaveOccurred())

		publicKey = &privateKey.PublicKey
		key = &certstrappkix.Key{
			Private: privateKey,
			Public:  publicKey,
		}

		ca, err = certstrappkix.CreateCertificateAuthority(caKey, "Cloud Foundry", 2, "Cloud Foundry", "USA", "CA", "San Francisco", "BOSH Bootloader")
		Expect(err).NotTo(HaveOccurred())

		exportedCA, err = ca.Export()
		Expect(err).NotTo(HaveOccurred())

		csr, err = certstrappkix.CreateCertificateSigningRequest(key, "Cloud Foundry", []net.IP{net.ParseIP("127.0.0.1")}, nil, "Cloud Foundry", "USA", "CA", "San Francisco", "127.0.0.1")
		Expect(err).NotTo(HaveOccurred())

		signedCert, err = certstrappkix.CreateCertificateHost(ca, caKey, csr, 2)
		Expect(err).NotTo(HaveOccurred())

		exportedCert, err = signedCert.Export()
		Expect(err).NotTo(HaveOccurred())

		fakeCertstrapPKIX.CreateCertificateAuthorityCall.Returns.Certificate = ca
		fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Returns.CertificateSigningRequest = csr
		fakeCertstrapPKIX.CreateCertificateHostCall.Returns.Certificate = signedCert

		fakePrivateKeyGenerator.GenerateKeyCall.Stub = func() (*rsa.PrivateKey, error) {
			if fakePrivateKeyGenerator.GenerateKeyCall.CallCount == 0 {
				return caPrivateKey, nil
			}

			return privateKey, nil
		}
	})

	Describe("Generate", func() {
		It("generates an SSL certificate signed by generated CA", func() {
			generatedKeyPair, err := generator.Generate("BOSH Bootloader", "127.0.0.1")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakePrivateKeyGenerator.GenerateKeyCall.CallCount).To(Equal(2))
			Expect(fakePrivateKeyGenerator.GenerateKeyCall.Receives[0].Random).To(Equal(rand.Reader))
			Expect(fakePrivateKeyGenerator.GenerateKeyCall.Receives[0].Bits).To(Equal(2048))
			Expect(fakePrivateKeyGenerator.GenerateKeyCall.Receives[1].Random).To(Equal(rand.Reader))
			Expect(fakePrivateKeyGenerator.GenerateKeyCall.Receives[1].Bits).To(Equal(2048))

			Expect(fakeCertstrapPKIX.CreateCertificateAuthorityCall.CallCount).To(Equal(1))
			Expect(fakeCertstrapPKIX.CreateCertificateAuthorityCall.Receives.Key.Private).To(Equal(caPrivateKey))
			Expect(fakeCertstrapPKIX.CreateCertificateAuthorityCall.Receives.Key.Public).To(Equal(caPublicKey))
			Expect(fakeCertstrapPKIX.CreateCertificateAuthorityCall.Receives.OrganizationalUnit).To(Equal("Cloud Foundry"))
			Expect(fakeCertstrapPKIX.CreateCertificateAuthorityCall.Receives.Years).To(Equal(2))
			Expect(fakeCertstrapPKIX.CreateCertificateAuthorityCall.Receives.Organization).To(Equal("Cloud Foundry"))
			Expect(fakeCertstrapPKIX.CreateCertificateAuthorityCall.Receives.Country).To(Equal("USA"))
			Expect(fakeCertstrapPKIX.CreateCertificateAuthorityCall.Receives.Province).To(Equal("CA"))
			Expect(fakeCertstrapPKIX.CreateCertificateAuthorityCall.Receives.Locality).To(Equal("San Francisco"))
			Expect(fakeCertstrapPKIX.CreateCertificateAuthorityCall.Receives.CommonName).To(Equal("BOSH Bootloader"))

			Expect(fakeCertstrapPKIX.CreateCertificateSigningRequestCall.CallCount).To(Equal(1))
			Expect(fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Receives.Key.Private).To(Equal(privateKey))
			Expect(fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Receives.Key.Public).To(Equal(publicKey))
			Expect(fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Receives.OrganizationalUnit).To(Equal("Cloud Foundry"))
			Expect(fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Receives.IpList).To(Equal([]net.IP{net.ParseIP("127.0.0.1")}))
			Expect(fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Receives.DomainList).To(BeNil())
			Expect(fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Receives.Organization).To(Equal("Cloud Foundry"))
			Expect(fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Receives.Country).To(Equal("USA"))
			Expect(fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Receives.Province).To(Equal("CA"))
			Expect(fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Receives.Locality).To(Equal("San Francisco"))
			Expect(fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Receives.CommonName).To(Equal("127.0.0.1"))

			Expect(fakeCertstrapPKIX.CreateCertificateHostCall.CallCount).To(Equal(1))
			Expect(fakeCertstrapPKIX.CreateCertificateHostCall.Receives.CrtAuth).To(Equal(ca))
			Expect(fakeCertstrapPKIX.CreateCertificateHostCall.Receives.KeyAuth.Private).To(Equal(caPrivateKey))
			Expect(fakeCertstrapPKIX.CreateCertificateHostCall.Receives.KeyAuth.Public).To(Equal(caPublicKey))
			Expect(fakeCertstrapPKIX.CreateCertificateHostCall.Receives.Csr).To(Equal(csr))
			Expect(fakeCertstrapPKIX.CreateCertificateHostCall.Receives.Years).To(Equal(2))

			Expect(generatedKeyPair.CA).To(Equal(exportedCA))
			Expect(generatedKeyPair.Certificate).To(Equal(exportedCert))
			Expect(strings.TrimSpace(string(generatedKeyPair.PrivateKey))).To(Equal(privateKeyPEM))
		})

		Context("failure cases", func() {
			Context("when private key generation fails for CA", func() {
				It("returns error", func() {
					fakePrivateKeyGenerator.GenerateKeyCall.Stub = func() (*rsa.PrivateKey, error) {
						return nil, errors.New("private key generation failed for ca")
					}

					_, err := generator.Generate("", "127.0.0.1")
					Expect(err).To(MatchError("private key generation failed for ca"))
				})
			})

			Context("when private key generation fails for certificate", func() {
				It("returns error", func() {
					fakePrivateKeyGenerator.GenerateKeyCall.Stub = func() (*rsa.PrivateKey, error) {
						if fakePrivateKeyGenerator.GenerateKeyCall.CallCount != 0 {
							return nil, errors.New("private key generation failed for certificate")
						}
						return privateKey, nil
					}

					_, err := generator.Generate("", "127.0.0.1")
					Expect(err).To(MatchError("private key generation failed for certificate"))
				})
			})

			It("errors when create certificate authority fails", func() {
				fakeCertstrapPKIX.CreateCertificateAuthorityCall.Returns.Error = errors.New("create certificate authority failed")

				_, err := generator.Generate("", "127.0.0.1")
				Expect(err).To(MatchError("create certificate authority failed"))
			})

			It("errors when create certificate signing request fails", func() {
				fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Returns.Error = errors.New("create certificate signing request failed")

				_, err := generator.Generate("", "127.0.0.1")
				Expect(err).To(MatchError("create certificate signing request failed"))
			})

			It("errors when create certificate host fails", func() {
				fakeCertstrapPKIX.CreateCertificateHostCall.Returns.Error = errors.New("could not generate certificate host")

				_, err := generator.Generate("", "127.0.0.1")
				Expect(err).To(MatchError("could not generate certificate host"))
			})
		})
	})
})
