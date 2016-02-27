package ec2

import (
	goaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type KeyPairUploader struct {
}

func NewKeyPairUploader() KeyPairUploader {
	return KeyPairUploader{}
}

func (k KeyPairUploader) Upload(client Session, keypair KeyPair) error {
	_, err := client.ImportKeyPair(&ec2.ImportKeyPairInput{
		KeyName:           goaws.String(keypair.Name),
		PublicKeyMaterial: keypair.PublicKey,
	})
	if err != nil {
		return err
	}

	return nil
}
