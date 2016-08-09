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
			fakeCertstrapPKIX.NewCertificateFromPEM,
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
		fakeCertstrapPKIX.NewCertificateFromPEMCall.Returns.Certificate = ca
	})

	Describe("GenerateCA", func() {
		BeforeEach(func() {
			fakePrivateKeyGenerator.GenerateKeyCall.Returns.PrivateKey = caPrivateKey
		})

		It("generates an SSL CA certificate", func() {
			generatedCA, err := generator.GenerateCA("BOSH Bootloader")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakePrivateKeyGenerator.GenerateKeyCall.CallCount).To(Equal(1))
			Expect(fakePrivateKeyGenerator.GenerateKeyCall.Receives.Random).To(Equal(rand.Reader))
			Expect(fakePrivateKeyGenerator.GenerateKeyCall.Receives.Bits).To(Equal(2048))

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

			Expect(generatedCA.CA).To(Equal(exportedCA))
			Expect(strings.TrimSpace(string(generatedCA.PrivateKey))).To(Equal(caPrivateKeyPEM))
		})

		Context("failure cases", func() {
			It("errors when private key generation fails", func() {
				fakePrivateKeyGenerator.GenerateKeyCall.Returns.Error = errors.New("private key generation failed")

				_, err := generator.GenerateCA("BOSH Bootloader")
				Expect(err).To(MatchError("private key generation failed"))
			})

			It("errors when create certificate authority fails", func() {
				fakeCertstrapPKIX.CreateCertificateAuthorityCall.Returns.Error = errors.New("create certificate authority failed")

				_, err := generator.GenerateCA("BOSH Bootloader")
				Expect(err).To(MatchError("create certificate authority failed"))
			})
		})
	})

	Describe("Generate", func() {
		var (
			generatedCA ssl.CAData
		)

		BeforeEach(func() {
			var err error
			fakePrivateKeyGenerator.GenerateKeyCall.Returns.PrivateKey = caPrivateKey
			generatedCA, err = generator.GenerateCA("BOSH Bootloader")
			Expect(err).NotTo(HaveOccurred())

			fakePrivateKeyGenerator.GenerateKeyCall.Returns.PrivateKey = privateKey
		})

		It("generates an SSL certificate signed by generated CA", func() {
			generatedKeyPair, err := generator.Generate(generatedCA, "127.0.0.1")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakePrivateKeyGenerator.GenerateKeyCall.CallCount).To(Equal(2))
			Expect(fakePrivateKeyGenerator.GenerateKeyCall.Receives.Random).To(Equal(rand.Reader))
			Expect(fakePrivateKeyGenerator.GenerateKeyCall.Receives.Bits).To(Equal(2048))
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

			csr := fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Returns.CertificateSigningRequest

			Expect(fakeCertstrapPKIX.CreateCertificateHostCall.Receives.CrtAuth).To(Equal(ca))
			Expect(fakeCertstrapPKIX.CreateCertificateHostCall.Receives.KeyAuth.Private).To(Equal(caPrivateKey))
			Expect(fakeCertstrapPKIX.CreateCertificateHostCall.Receives.KeyAuth.Public).To(Equal(caPublicKey))
			Expect(fakeCertstrapPKIX.CreateCertificateHostCall.Receives.Csr).To(Equal(csr))
			Expect(fakeCertstrapPKIX.CreateCertificateHostCall.Receives.Years).To(Equal(2))

			Expect(generatedKeyPair.CA).To(Equal(generatedCA.CA))
			Expect(generatedKeyPair.Certificate).To(Equal(exportedCert))

			pkeyDER, rest := pem.Decode(generatedKeyPair.PrivateKey)
			Expect(rest).To(HaveLen(0))

			decodedPrivateKey, err := x509.ParsePKCS1PrivateKey(pkeyDER.Bytes)
			Expect(err).NotTo(HaveOccurred())

			err = decodedPrivateKey.Validate()
			Expect(err).NotTo(HaveOccurred())

			Expect(privateKey).To(Equal(decodedPrivateKey))
		})

		Context("failure cases", func() {
			It("errors when private key generation fails", func() {
				fakePrivateKeyGenerator.GenerateKeyCall.Returns.Error = errors.New("private key generation failed")

				_, err := generator.Generate(generatedCA, "127.0.0.1")
				Expect(err).To(MatchError("private key generation failed"))
			})

			It("errors when createCertificateSigningRequest fails", func() {
				fakeCertstrapPKIX.CreateCertificateSigningRequestCall.Returns.Error = errors.New("create certificate signing request failed")

				_, err := generator.Generate(generatedCA, "127.0.0.1")
				Expect(err).To(MatchError("create certificate signing request failed"))
			})

			It("errors when NewCertFromPem fails", func() {
				fakeCertstrapPKIX.NewCertificateFromPEMCall.Returns.Error = errors.New("new certificate from pem failed")

				_, err := generator.Generate(ssl.CAData{}, "127.0.0.1")
				Expect(err).To(MatchError("new certificate from pem failed"))
			})

			It("errors when create certificate host fails", func() {
				fakeCertstrapPKIX.CreateCertificateHostCall.Returns.Error = errors.New("could not generate certificate host")

				_, err := generator.Generate(generatedCA, "127.0.0.1")
				Expect(err).To(MatchError("could not generate certificate host"))
			})

			It("errors when create certificate host fails", func() {
				_, err := generator.Generate(ssl.CAData{
					PrivateKey: []byte("-----BEGIN RSA PRIVATE KEY-----\n-----END RSA PRIVATE KEY-----"),
				}, "127.0.0.1")
				Expect(err).To(MatchError(ContainSubstring("sequence truncated")))
			})
		})
	})
})
