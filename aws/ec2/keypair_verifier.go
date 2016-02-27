package ec2

import (
	"crypto/md5"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

type KeyPairVerifier struct{}

func (v KeyPairVerifier) Verify(fingerprint string, pemData []byte) error {

	pem, _ := pem.Decode(pemData)
	if pem == nil {
		return errors.New("the local keypair does not contain a valid PEM encoded private key, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance.")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(pem.Bytes)
	if err != nil {
		return errors.New("the local keypair does not contain a valid rsa private key, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance.")
	}

	pkeyBytes, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		return err
	}

	var parts []string
	for _, c := range md5.Sum(pkeyBytes) {
		parts = append(parts, fmt.Sprintf("%02x", c))
	}

	if fingerprint != strings.Join(parts, ":") {
		return errors.New("the local keypair fingerprint does not match the keypair fingerprint on AWS")
	}

	return nil
}
