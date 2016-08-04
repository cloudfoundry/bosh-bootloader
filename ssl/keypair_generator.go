package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"time"
)

type clock func() time.Time
type keyGenerator func(io.Reader, int) (*rsa.PrivateKey, error)
type certCreator func(rand io.Reader, template, parent *x509.Certificate, pub, priv interface{}) ([]byte, error)
type certParser func(asn1Data []byte) ([]*x509.Certificate, error)

type KeyPairGenerator struct {
	getTime           clock
	generateKey       keyGenerator
	createCertificate certCreator
	parseCertificates certParser
}

func NewKeyPairGenerator(clock clock, keyGenerator keyGenerator, certCreator certCreator, certParser certParser) KeyPairGenerator {
	return KeyPairGenerator{
		getTime:           clock,
		generateKey:       keyGenerator,
		createCertificate: certCreator,
		parseCertificates: certParser,
	}
}

func (g KeyPairGenerator) GenerateCA(commonName string) ([]byte, error) {
	now := g.getTime()
	template := x509.Certificate{
		SerialNumber: big.NewInt(1234),
		Subject: pkix.Name{
			Country:      []string{"USA"},
			Organization: []string{"Cloud Foundry"},
			CommonName:   commonName,
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(2, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA: true,
	}

	privatekey, err := g.generateKey(rand.Reader, 2048)
	if err != nil {
		return []byte{}, err
	}

	cert, err := g.createCertificate(rand.Reader, &template, &template, &privatekey.PublicKey, privatekey)
	if err != nil {
		return []byte{}, err
	}

	return cert, nil
}

func (g KeyPairGenerator) Generate(ca []byte, commonName string) (KeyPair, error) {
	now := g.getTime()
	template := x509.Certificate{
		SerialNumber: big.NewInt(1234),
		Subject: pkix.Name{
			Country:      []string{"USA"},
			Organization: []string{"Cloud Foundry"},
			CommonName:   commonName,
		},
		NotBefore:   now,
		NotAfter:    now.AddDate(2, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageKeyEncipherment,
		IPAddresses: []net.IP{
			net.ParseIP(commonName),
		},
	}

	privatekey, err := g.generateKey(rand.Reader, 2048)
	if err != nil {
		return KeyPair{}, err
	}

	parsedCerts, err := g.parseCertificates(ca)
	if err != nil {
		return KeyPair{}, err
	}

	caCert := parsedCerts[0]

	cert, err := g.createCertificate(rand.Reader, &template, caCert, &privatekey.PublicKey, privatekey)
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		CA: pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: ca,
		}),
		Certificate: pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert,
		}),
		PrivateKey: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privatekey),
		}),
	}, nil
}
