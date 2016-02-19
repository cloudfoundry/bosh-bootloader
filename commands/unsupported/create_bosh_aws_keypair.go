package unsupported

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/state"
)

type keypairGenerator interface {
	Generate() (ec2.Keypair, error)
}

type keypairUploader interface {
	Upload(ec2.Session, ec2.Keypair) error
}

type stateStore interface {
	Set(directory string, s state.State) error
	Get(director string) (state.State, error)
}

type sessionProvider interface {
	Session(ec2.Config) (ec2.Session, error)
}

type CreateBoshAWSKeypair struct {
	generator keypairGenerator
	uploader  keypairUploader
	provider  sessionProvider
	store     stateStore
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

	config, err := getConfig(c.store, globalFlags.StateDir, ec2.Config{
		AccessKeyID:      globalFlags.AWSAccessKeyID,
		SecretAccessKey:  globalFlags.AWSSecretAccessKey,
		Region:           globalFlags.AWSRegion,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return err
	}

	session, err := c.provider.Session(config)
	if err != nil {
		return err
	}

	err = c.uploader.Upload(session, keypair)
	if err != nil {
		return err
	}

	err = c.store.Set(globalFlags.StateDir, state.State{
		AWS: state.AWS{
			AccessKeyID:     config.AccessKeyID,
			SecretAccessKey: config.SecretAccessKey,
			Region:          config.Region,
		},
		KeyPair: &state.KeyPair{
			Name:       keypair.Name,
			PublicKey:  string(keypair.PublicKey),
			PrivateKey: string(keypair.PrivateKey),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func getConfig(store stateStore, dir string, config ec2.Config) (ec2.Config, error) {
	state, err := store.Get(dir)
	if err != nil {
		return config, err
	}

	if config.AccessKeyID == "" {
		config.AccessKeyID = state.AWS.AccessKeyID
	}

	if config.SecretAccessKey == "" {
		config.SecretAccessKey = state.AWS.SecretAccessKey
	}

	if config.Region == "" {
		config.Region = state.AWS.Region
	}

	return config, nil
}
