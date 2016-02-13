package ec2

import (
	"crypto/rsa"
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
	Name string
	Key  []byte
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

	return Keypair{
		Name: fmt.Sprintf("keypair-%s", uuid),
		Key:  ssh.MarshalAuthorizedKey(pub),
	}, nil
}
