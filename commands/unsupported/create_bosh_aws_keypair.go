package unsupported

import (
	"crypto/md5"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
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

func (c CreateBoshAWSKeypair) Execute(globalFlags commands.GlobalFlags, state storage.State) (storage.State, error) {
	session, err := c.provider.Session(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	if state.KeyPair != nil {
		keyInfo, err := c.retriever.Retrieve(session, state.KeyPair.Name)
		if err != nil {
			if err != ec2.KeyPairNotFound {
				return state, err
			}

			err := c.uploader.Upload(session, ec2.Keypair{
				Name:      state.KeyPair.Name,
				PublicKey: []byte(state.KeyPair.PublicKey),
			})
			if err != nil {
				return state, err
			}

			return state, nil
		}

		fingerprintMatches, err := verifyFingerprint(keyInfo.Fingerprint, []byte(state.KeyPair.PrivateKey))
		if err != nil {
			return state, err
		}

		if !fingerprintMatches {
			return state, NewIssue("the local keypair fingerprint does not match the keypair fingerprint on AWS")
		}

		return state, nil
	}

	keypair, err := c.generateAndUploadKeypair(session)
	if err != nil {
		return state, err
	}

	state.KeyPair = &storage.KeyPair{
		Name:       keypair.Name,
		PrivateKey: string(keypair.PrivateKey),
		PublicKey:  string(keypair.PublicKey),
	}

	return state, nil
}

func verifyFingerprint(awsFingerprint string, privateKeyPem []byte) (bool, error) {
	pem, _ := pem.Decode(privateKeyPem)
	if pem == nil {
		return false, NewIssue("the local keypair does not contain a valid PEM encoded private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(pem.Bytes)
	if err != nil {
		return false, NewIssue("the local keypair does not contain a valid rsa private key")
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
