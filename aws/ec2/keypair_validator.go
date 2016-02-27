package ec2

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type KeyPairValidator struct{}

func (v KeyPairValidator) Validate(pemData []byte) error {
	pem, _ := pem.Decode(pemData)
	if pem == nil {
		return errors.New("the local keypair does not contain a valid PEM encoded private key, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance.")
	}

	_, err := x509.ParsePKCS1PrivateKey(pem.Bytes)
	if err != nil {
		return errors.New("the local keypair does not contain a valid rsa private key, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance.")
	}

	return nil
}
