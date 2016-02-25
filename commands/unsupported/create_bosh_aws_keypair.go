package unsupported

import (
	"crypto/md5"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/state"
)

type keypairRetriever interface {
	Retrieve(session ec2.Session, name string) (ec2.KeyPairInfo, error)
}

type keypairGenerator interface {
	Generate() (ec2.Keypair, error)
}

type keypairUploader interface {
	Upload(ec2.Session, ec2.Keypair) error
}

type sessionProvider interface {
	Session(aws.Config) (ec2.Session, error)
}

type CreateBoshAWSKeypair struct {
	retriever keypairRetriever
	generator keypairGenerator
	uploader  keypairUploader
	provider  sessionProvider
}

func NewCreateBoshAWSKeypair(retriever keypairRetriever, generator keypairGenerator, uploader keypairUploader, provider sessionProvider) CreateBoshAWSKeypair {
	return CreateBoshAWSKeypair{
		retriever: retriever,
		generator: generator,
		uploader:  uploader,
		provider:  provider,
	}
}

func (c CreateBoshAWSKeypair) Execute(globalFlags commands.GlobalFlags, s state.State) (state.State, error) {
	session, err := c.provider.Session(aws.Config{
		AccessKeyID:      s.AWS.AccessKeyID,
		SecretAccessKey:  s.AWS.SecretAccessKey,
		Region:           s.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return s, err
	}

	if s.KeyPair != nil {
		keyInfo, err := c.retriever.Retrieve(session, s.KeyPair.Name)
		if err != nil {
			if err != ec2.KeyPairNotFound {
				return s, err
			}

			err := c.uploader.Upload(session, ec2.Keypair{
				Name:      s.KeyPair.Name,
				PublicKey: []byte(s.KeyPair.PublicKey),
			})
			if err != nil {
				return s, err
			}

			return s, nil
		}

		fingerprintMatches, err := verifyFingerprint(keyInfo.Fingerprint, []byte(s.KeyPair.PrivateKey))
		if err != nil {
			return s, err
		}

		if !fingerprintMatches {
			return s, errors.New("the local keypair fingerprint does not match the keypair fingerprint on AWS, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance.")
		}

		return s, nil
	}

	keypair, err := c.generateAndUploadKeypair(session)
	if err != nil {
		return s, err
	}

	s.KeyPair = &state.KeyPair{
		Name:       keypair.Name,
		PrivateKey: string(keypair.PrivateKey),
		PublicKey:  string(keypair.PublicKey),
	}

	return s, nil
}

func verifyFingerprint(awsFingerprint string, privateKeyPem []byte) (bool, error) {
	pem, _ := pem.Decode(privateKeyPem)
	if pem == nil {
		return false, errors.New("the local keypair does not contain a valid PEM encoded private key, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance.")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(pem.Bytes)
	if err != nil {
		return false, errors.New("the local keypair does not contain a valid rsa private key, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance.")
	}

	pkeyBytes, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		return false, err
	}

	var fingerprint string
	for _, c := range md5.Sum(pkeyBytes) {
		fingerprint += fmt.Sprintf(":%02x", c)
	}

	fingerprint = strings.TrimPrefix(fingerprint, ":")

	if awsFingerprint != fingerprint {
		return false, nil
	}

	return true, nil
}

func (c CreateBoshAWSKeypair) generateAndUploadKeypair(session ec2.Session) (ec2.Keypair, error) {
	keypair, err := c.generator.Generate()
	if err != nil {
		return ec2.Keypair{}, err
	}

	err = c.uploader.Upload(session, keypair)
	if err != nil {
		return ec2.Keypair{}, err
	}

	return keypair, nil
}
