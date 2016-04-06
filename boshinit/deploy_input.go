package boshinit

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

const USERNAME_PREFIX = "user-"
const USERNAME_LENGTH = 7
const PASSWORD_PREFIX = "p-"
const PASSWORD_LENGTH = 15

type DeployInput struct {
	DirectorUsername            string
	DirectorPassword            string
	State                       State
	InfrastructureConfiguration InfrastructureConfiguration
	SSLKeyPair                  ssl.KeyPair
	EC2KeyPair                  ec2.KeyPair
	Credentials                 map[string]string
}

type InfrastructureConfiguration struct {
	AWSRegion        string
	SubnetID         string
	AvailabilityZone string
	ElasticIP        string
	AccessKeyID      string
	SecretAccessKey  string
	SecurityGroup    string
}

type DeployOutput struct {
	Credentials        map[string]string
	BOSHInitState      State
	DirectorSSLKeyPair ssl.KeyPair
	BOSHInitManifest   string
}

type stringGenerator interface {
	Generate(prefix string, length int) (string, error)
}

func NewDeployInput(state storage.State, infrastructureConfiguration InfrastructureConfiguration, stringGenerator stringGenerator) (DeployInput, error) {
	deployInput := DeployInput{
		State: map[string]interface{}{},
		InfrastructureConfiguration: infrastructureConfiguration,
		SSLKeyPair:                  ssl.KeyPair{},
		EC2KeyPair:                  ec2.KeyPair{},
	}

	if !state.KeyPair.IsEmpty() {
		deployInput.EC2KeyPair.Name = state.KeyPair.Name
		deployInput.EC2KeyPair.PrivateKey = state.KeyPair.PrivateKey
		deployInput.EC2KeyPair.PublicKey = state.KeyPair.PublicKey
	}

	if !state.BOSH.IsEmpty() {
		deployInput.DirectorUsername = state.BOSH.DirectorUsername
		deployInput.DirectorPassword = state.BOSH.DirectorPassword
		deployInput.State = state.BOSH.State
		deployInput.Credentials = state.BOSH.Credentials
		deployInput.SSLKeyPair.Certificate = []byte(state.BOSH.DirectorSSLCertificate)
		deployInput.SSLKeyPair.PrivateKey = []byte(state.BOSH.DirectorSSLPrivateKey)
	}

	if deployInput.DirectorUsername == "" {
		var err error
		if deployInput.DirectorUsername, err = stringGenerator.Generate(USERNAME_PREFIX, USERNAME_LENGTH); err != nil {
			return DeployInput{}, err
		}
	}

	if deployInput.DirectorPassword == "" {
		var err error
		if deployInput.DirectorPassword, err = stringGenerator.Generate(PASSWORD_PREFIX, PASSWORD_LENGTH); err != nil {
			return DeployInput{}, err
		}
	}

	return deployInput, nil
}
