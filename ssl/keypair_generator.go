package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net"

	certstrappkix "github.com/square/certstrap/pkix"
)

type keyGenerator func(io.Reader, int) (*rsa.PrivateKey, error)
type createCertificateAuthority func(key *certstrappkix.Key, organizationalUnit string, years int, organization string, country string, province string, locality string, commonName string) (*certstrappkix.Certificate, error)
type createCertificateSigningRequest func(key *certstrappkix.Key, organizationalUnit string, ipList []net.IP, domainList []string, organization string, country string, province string, locality string, commonName string) (*certstrappkix.CertificateSigningRequest, error)
type createCertificateHost func(crtAuth *certstrappkix.Certificate, keyAuth *certstrappkix.Key, csr *certstrappkix.CertificateSigningRequest, years int) (*certstrappkix.Certificate, error)
type newCertificateFromPEM func(pemCertificate []byte) (*certstrappkix.Certificate, error)

type KeyPairGenerator struct {
	generateKey                     keyGenerator
	createCertificateAuthority      createCertificateAuthority
	createCertificateSigningRequest createCertificateSigningRequest
	createCertificateHost           createCertificateHost
	newCertificateFromPEM           newCertificateFromPEM
}

type CAData struct {
	CA         []byte
	PrivateKey []byte
}

func NewKeyPairGenerator(
	keyGenerator keyGenerator,
	createCertificateAuthority createCertificateAuthority,
	createCertificateSigningRequest createCertificateSigningRequest,
	createCertificateHost createCertificateHost,
	newCertificateFromPEM newCertificateFromPEM,
) KeyPairGenerator {
	return KeyPairGenerator{
		generateKey:                     keyGenerator,
		createCertificateAuthority:      createCertificateAuthority,
		createCertificateSigningRequest: createCertificateSigningRequest,
		createCertificateHost:           createCertificateHost,
		newCertificateFromPEM:           newCertificateFromPEM,
	}
}

func (g KeyPairGenerator) GenerateCA(commonName string) (CAData, error) {
	privateKey, err := g.generateKey(rand.Reader, 2048)
	if err != nil {
		return CAData{}, err
	}

	key := certstrappkix.NewKey(&privateKey.PublicKey, privateKey)

	cert, err := g.createCertificateAuthority(key, "Cloud Foundry", 2, "Cloud Foundry", "USA", "CA", "San Francisco", commonName)
	if err != nil {
		return CAData{}, err
	}

	pemCertificate, err := cert.Export()
	if err != nil {
		return CAData{}, err
	}

	return CAData{
		CA: pemCertificate,
		PrivateKey: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}),
	}, nil
}

func (g KeyPairGenerator) Generate(caData CAData, commonName string) (KeyPair, error) {
	privateKey, err := g.generateKey(rand.Reader, 2048)
	if err != nil {
		return KeyPair{}, err
	}

	certKey := certstrappkix.Key{
		Private: privateKey,
		Public:  &privateKey.PublicKey,
	}

	ipList := []net.IP{
		net.ParseIP(commonName),
	}

	csr, err := g.createCertificateSigningRequest(&certKey, "Cloud Foundry", ipList, nil, "Cloud Foundry",
		"USA", "CA", "San Francisco", commonName)
	if err != nil {
		return KeyPair{}, err
	}

	caCertificate, err := g.newCertificateFromPEM(caData.CA)
	if err != nil {
		return KeyPair{}, err
	}

	caPrivateKeyDER, _ := pem.Decode(caData.PrivateKey)

	decodedCAPrivateKey, err := x509.ParsePKCS1PrivateKey(caPrivateKeyDER.Bytes)
	if err != nil {
		return KeyPair{}, err
	}

	caKey := certstrappkix.Key{
		Private: decodedCAPrivateKey,
		Public:  &decodedCAPrivateKey.PublicKey,
	}

	signedCert, err := g.createCertificateHost(caCertificate, &caKey, csr, 2)
	if err != nil {
		return KeyPair{}, err
	}

	pemCertificate, err := signedCert.Export()
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		CA:          caData.CA,
		Certificate: pemCertificate,
		PrivateKey: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}),
	}, nil
}
