package unsupported

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
)

type keypairGenerator interface {
	Generate() (ec2.Keypair, error)
}

type keypairUploader interface {
	Upload(ec2.Session, ec2.Keypair) error
}

type stateStore interface {
	Merge(directory string, stateMap map[string]interface{}) error
	GetString(directory, key string) (value string, present bool, err error)
}

type CreateBoshAWSKeypair struct {
	generator keypairGenerator
	uploader  keypairUploader
	provider  sessionProvider
	store     stateStore
}

type sessionProvider interface {
	Session(ec2.Config) ec2.Session
}

func NewCreateBoshAWSKeypair(generator keypairGenerator, uploader keypairUploader, provider sessionProvider, store stateStore) CreateBoshAWSKeypair {
	return CreateBoshAWSKeypair{
		generator: generator,
		uploader:  uploader,
		provider:  provider,
		store:     store,
	}
}

func (c CreateBoshAWSKeypair) Execute(globalFlags commands.GlobalFlags) error {
	keypair, err := c.generator.Generate()
	if err != nil {
		return err
	}

	config := ec2.Config{
		AccessKeyID:      globalFlags.AWSAccessKeyID,
		SecretAccessKey:  globalFlags.AWSSecretAccessKey,
		Region:           globalFlags.AWSRegion,
		EndpointOverride: globalFlags.EndpointOverride,
	}

	if config.AccessKeyID == "" {
		value, ok, err := c.store.GetString(globalFlags.StateDir, "aws-access-key-id")
		if err != nil {
			return err
		}
		if ok {
			config.AccessKeyID = value
		}
	}

	if config.SecretAccessKey == "" {
		value, ok, err := c.store.GetString(globalFlags.StateDir, "aws-secret-access-key")
		if err != nil {
			return err
		}
		if ok {
			config.SecretAccessKey = value
		}
	}

	if config.Region == "" {
		value, ok, err := c.store.GetString(globalFlags.StateDir, "aws-region")
		if err != nil {
			return err
		}
		if ok {
			config.Region = value
		}
	}

	if config.AccessKeyID == "" || config.SecretAccessKey == "" || config.Region == "" {
		return errors.New("aws credentials must be provided")
	}

	session := c.provider.Session(config)
	err = c.uploader.Upload(session, keypair)
	if err != nil {
		return err
	}

	err = c.store.Merge(globalFlags.StateDir, map[string]interface{}{
		"aws-access-key-id":     config.AccessKeyID,
		"aws-secret-access-key": config.SecretAccessKey,
		"aws-region":            config.Region,
	})
	if err != nil {
		return err
	}

	return nil
}
