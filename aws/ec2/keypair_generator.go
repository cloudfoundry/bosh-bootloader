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

type KeypairGenerator struct {
	generateUUID         uuidGenerator
	random               io.Reader
	generateRSAKey       rsaKeyGenerator
	generateSSHPublicKey sshPublicKeyGenerator
}

type Keypair struct {
	Name       string
	PublicKey  []byte
	PrivateKey []byte
}

func NewKeypairGenerator(random io.Reader, generateUUID uuidGenerator, generateRSAKey rsaKeyGenerator, generateSSHPublicKey sshPublicKeyGenerator) KeypairGenerator {
	return KeypairGenerator{
		random:               random,
		generateUUID:         generateUUID,
		generateRSAKey:       generateRSAKey,
		generateSSHPublicKey: generateSSHPublicKey,
	}
}

func (k KeypairGenerator) Generate() (Keypair, error) {
	rsakey, err := k.generateRSAKey(k.random, 2048)
	if err != nil {
		return Keypair{}, err
	}

	pub, err := k.generateSSHPublicKey(rsakey.Public())
	if err != nil {
		return Keypair{}, err
	}

	uuid, err := k.generateUUID()
	if err != nil {
		return Keypair{}, err
	}

	privateKey := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(rsakey),
		},
	)

	return Keypair{
		Name:       fmt.Sprintf("keypair-%s", uuid),
		PublicKey:  ssh.MarshalAuthorizedKey(pub),
		PrivateKey: privateKey,
	}, nil
}
