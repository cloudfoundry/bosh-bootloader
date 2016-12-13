package gcp

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"strings"

	compute "google.golang.org/api/compute/v1"

	"github.com/cloudfoundry/bosh-bootloader/storage"

	"golang.org/x/crypto/ssh"
)

type KeyPairUpdater struct {
	random                io.Reader
	rsaKeyGenerator       rsaKeyGenerator
	sshPublicKeyGenerator sshPublicKeyGenerator
	clientProvider        clientProvider
	logger                logger
}

type rsaKeyGenerator func(io.Reader, int) (*rsa.PrivateKey, error)
type sshPublicKeyGenerator func(interface{}) (ssh.PublicKey, error)

type clientProvider interface {
	Client() Client
}

type logger interface {
	Step(string, ...interface{})
}

func NewKeyPairUpdater(random io.Reader, generateRSAKey rsaKeyGenerator, generateSSHPublicKey sshPublicKeyGenerator, clientProvider clientProvider, logger logger) KeyPairUpdater {
	return KeyPairUpdater{
		random:                random,
		rsaKeyGenerator:       generateRSAKey,
		sshPublicKeyGenerator: generateSSHPublicKey,
		clientProvider:        clientProvider,
		logger:                logger,
	}
}

func (k KeyPairUpdater) Update() (storage.KeyPair, error) {
	privateKey, publicKey, err := k.createKeyPair()
	if err != nil {
		return storage.KeyPair{}, err
	}

	client := k.clientProvider.Client()
	project, err := client.GetProject()
	if err != nil {
		return storage.KeyPair{}, err
	}

	sshKeyItemValue := fmt.Sprintf("vcap:%s vcap", strings.TrimSpace(publicKey))

	var updated bool
	for i, item := range project.CommonInstanceMetadata.Items {
		if item.Key == "sshKeys" {
			sshKeys := strings.Split(*item.Value, "\n")
			sshKeys = append(sshKeys, sshKeyItemValue)

			newValue := strings.Join(sshKeys, "\n")
			project.CommonInstanceMetadata.Items[i].Value = &newValue
			updated = true
			k.logger.Step("appending new ssh-keys for the project %q", client.ProjectID())
			break
		}
	}

	if !updated {
		k.logger.Step("Creating new ssh-keys for the project %q", client.ProjectID())
		sshKeyItem := &compute.MetadataItems{
			Key:   "sshKeys",
			Value: &sshKeyItemValue,
		}

		project.CommonInstanceMetadata.Items = append(project.CommonInstanceMetadata.Items, sshKeyItem)
	}

	_, err = client.SetCommonInstanceMetadata(project.CommonInstanceMetadata)
	if err != nil {
		return storage.KeyPair{}, err
	}

	return storage.KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

func (keyPairUpdater KeyPairUpdater) createKeyPair() (string, string, error) {
	rsaKey, err := keyPairUpdater.rsaKeyGenerator(keyPairUpdater.random, 2048)
	if err != nil {
		return "", "", err
	}

	publicKey, err := keyPairUpdater.sshPublicKeyGenerator(rsaKey.Public())
	if err != nil {
		return "", "", err
	}

	rawPublicKey := string(ssh.MarshalAuthorizedKey(publicKey))
	rawPublicKey = strings.TrimSuffix(rawPublicKey, "\n")

	privateKey := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
		},
	)

	return string(privateKey), rawPublicKey, nil
}
