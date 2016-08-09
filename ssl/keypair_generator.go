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

type KeyPairGenerator struct {
	generateKey                     keyGenerator
	createCertificateAuthority      createCertificateAuthority
	createCertificateSigningRequest createCertificateSigningRequest
	createCertificateHost           createCertificateHost
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
) KeyPairGenerator {
	return KeyPairGenerator{
		generateKey:                     keyGenerator,
		createCertificateAuthority:      createCertificateAuthority,
		createCertificateSigningRequest: createCertificateSigningRequest,
		createCertificateHost:           createCertificateHost,
	}
}

func (g KeyPairGenerator) Generate(caCommonName, commonName string) (KeyPair, error) {
	caPrivateKey, err := g.generateKey(rand.Reader, 2048)
	if err != nil {
		return KeyPair{}, err
	}
	caKey := certstrappkix.NewKey(&caPrivateKey.PublicKey, caPrivateKey)

	caCertificate, err := g.createCertificateAuthority(caKey, "Cloud Foundry", 2, "Cloud Foundry", "USA", "CA",
		"San Francisco", caCommonName)
	if err != nil {
		return KeyPair{}, err
	}

	certPrivateKey, err := g.generateKey(rand.Reader, 2048)
	if err != nil {
		return KeyPair{}, err
	}
	certKey := certstrappkix.NewKey(&certPrivateKey.PublicKey, certPrivateKey)

	ipList := []net.IP{
		net.ParseIP(commonName),
	}

	csr, err := g.createCertificateSigningRequest(certKey, "Cloud Foundry", ipList, nil, "Cloud Foundry",
		"USA", "CA", "San Francisco", commonName)
	if err != nil {
		return KeyPair{}, err
	}

	certificate, err := g.createCertificateHost(caCertificate, caKey, csr, 2)
	if err != nil {
		return KeyPair{}, err
	}

	pemCA, err := caCertificate.Export()
	if err != nil {
		return KeyPair{}, err
	}

	pemCertificate, err := certificate.Export()
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		CA:          pemCA,
		Certificate: pemCertificate,
		PrivateKey: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(certPrivateKey),
		}),
	}, nil
}
