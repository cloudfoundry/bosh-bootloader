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

type KeyPairGenerator struct {
	getTime           clock
	generateKey       keyGenerator
	createCertificate certCreator
}

func NewKeyPairGenerator(clock clock, keyGenerator keyGenerator, certCreator certCreator) KeyPairGenerator {
	return KeyPairGenerator{
		getTime:           clock,
		generateKey:       keyGenerator,
		createCertificate: certCreator,
	}
}

func (g KeyPairGenerator) Generate(commonName string) (KeyPair, error) {
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

	cert, err := g.createCertificate(rand.Reader, &template, &template, &privatekey.PublicKey, privatekey)
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
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
