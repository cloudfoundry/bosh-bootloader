package gcp

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"

	"golang.org/x/crypto/ssh"
)

type KeyPairCreator struct {
	random                io.Reader
	rsaKeyGenerator       rsaKeyGenerator
	sshPublicKeyGenerator sshPublicKeyGenerator
}

type rsaKeyGenerator func(io.Reader, int) (*rsa.PrivateKey, error)
type sshPublicKeyGenerator func(interface{}) (ssh.PublicKey, error)

func NewKeyPairCreator(random io.Reader, generateRSAKey rsaKeyGenerator, generateSSHPublicKey sshPublicKeyGenerator) KeyPairCreator {
	return KeyPairCreator{
		random:                random,
		rsaKeyGenerator:       generateRSAKey,
		sshPublicKeyGenerator: generateSSHPublicKey,
	}
}

func (keyPairCreator KeyPairCreator) Create() (string, string, error) {
	rsaKey, err := keyPairCreator.rsaKeyGenerator(keyPairCreator.random, 2048)
	if err != nil {
		return "", "", err
	}

	publicKey, err := keyPairCreator.sshPublicKeyGenerator(rsaKey.Public())
	if err != nil {
		return "", "", err
	}

	privateKey := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
		},
	)

	return string(privateKey), string(ssh.MarshalAuthorizedKey(publicKey)), nil
}
