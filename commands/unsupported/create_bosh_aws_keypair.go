package unsupported

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
)

type keypairGenerator interface {
	Generate() (ec2.Keypair, error)
}

type keypairUploader interface {
	Upload(ec2.Session, ec2.Keypair) error
}

type CreateBoshAWSKeypair struct {
	generator keypairGenerator
	uploader  keypairUploader
	provider  sessionProvider
}

type sessionProvider interface {
	Session(ec2.Config) ec2.Session
}

func NewCreateBoshAWSKeypair(generator keypairGenerator, uploader keypairUploader, provider sessionProvider) CreateBoshAWSKeypair {
	return CreateBoshAWSKeypair{
		generator: generator,
		uploader:  uploader,
		provider:  provider,
	}
}

func (c CreateBoshAWSKeypair) Execute(globalFlags commands.GlobalFlags) error {
	keypair, err := c.generator.Generate()
	if err != nil {
		return err
	}

	session := c.provider.Session(ec2.Config{
		AccessKeyID:      globalFlags.AWSAccessKeyID,
		SecretAccessKey:  globalFlags.AWSSecretAccessKey,
		Region:           globalFlags.AWSRegion,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	err = c.uploader.Upload(session, keypair)
	if err != nil {
		return err
	}

	return nil
}
