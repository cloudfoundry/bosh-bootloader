package ssl

import (
	"crypto/x509"
	"encoding/pem"
)

type KeyPair struct {
	CA          []byte
	Certificate []byte
	PrivateKey  []byte
}

func (k KeyPair) IsEmpty() bool {
	return len(k.Certificate) == 0 || len(k.PrivateKey) == 0
}

func (k KeyPair) IsValidForIP(ip string) bool {
	if k.IsEmpty() {
		return false
	}

	block, _ := pem.Decode(k.Certificate)
	if block == nil {
		return false
	}

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}

	err = certificate.VerifyHostname(ip)
	return err == nil
}
