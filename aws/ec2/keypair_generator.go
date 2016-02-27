package ec2

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
)

type uuidGenerator func() (string, error)
type rsaKeyGenerator func(io.Reader, int) (*rsa.PrivateKey, error)
type sshPublicKeyGenerator func(interface{}) (ssh.PublicKey, error)

type KeyPairGenerator struct {
	generateUUID         uuidGenerator
	random               io.Reader
	generateRSAKey       rsaKeyGenerator
	generateSSHPublicKey sshPublicKeyGenerator
}


func NewKeyPairGenerator(random io.Reader, generateUUID uuidGenerator, generateRSAKey rsaKeyGenerator, generateSSHPublicKey sshPublicKeyGenerator) KeyPairGenerator {
	return KeyPairGenerator{
		random:               random,
		generateUUID:         generateUUID,
		generateRSAKey:       generateRSAKey,
		generateSSHPublicKey: generateSSHPublicKey,
	}
}

func (k KeyPairGenerator) Generate() (KeyPair, error) {
	rsakey, err := k.generateRSAKey(k.random, 2048)
	if err != nil {
		return KeyPair{}, err
	}

	pub, err := k.generateSSHPublicKey(rsakey.Public())
	if err != nil {
		return KeyPair{}, err
	}

	uuid, err := k.generateUUID()
	if err != nil {
		return KeyPair{}, err
	}

	privateKey := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(rsakey),
		},
	)

	return KeyPair{
		Name:       fmt.Sprintf("keypair-%s", uuid),
		PublicKey:  ssh.MarshalAuthorizedKey(pub),
		PrivateKey: privateKey,
	}, nil
}
