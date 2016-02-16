package ec2

import (
	goaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type KeypairUploader struct {
}

func NewKeypairUploader() KeypairUploader {
	return KeypairUploader{}
}

func (k KeypairUploader) Upload(importer Session, keypair Keypair) error {
	_, err := importer.ImportKeyPair(&ec2.ImportKeyPairInput{
		KeyName:           goaws.String(keypair.Name),
		PublicKeyMaterial: keypair.Key,
	})
	if err != nil {
		return err
	}

	return nil
}
