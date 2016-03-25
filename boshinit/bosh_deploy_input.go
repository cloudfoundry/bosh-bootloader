package boshinit

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

const USERNAME_PREFIX = "user-"
const USERNAME_LENGTH = 7
const PASSWORD_PREFIX = "p-"
const PASSWORD_LENGTH = 15

type BOSHDeployInput struct {
	DirectorUsername string
	DirectorPassword string
	State            State
	Stack            cloudformation.Stack
	AWSRegion        string
	SSLKeyPair       ssl.KeyPair
	EC2KeyPair       ec2.KeyPair
	Credentials      map[string]string
}

type stringGenerator interface {
	Generate(prefix string, length int) (string, error)
}

func NewBOSHDeployInput(state storage.State, stack cloudformation.Stack, stringGenerator stringGenerator) (BOSHDeployInput, error) {
	var err error
	boshDeployInput := BOSHDeployInput{
		State:      map[string]interface{}{},
		Stack:      stack,
		AWSRegion:  state.AWS.Region,
		SSLKeyPair: ssl.KeyPair{},
		EC2KeyPair: ec2.KeyPair{},
	}

	if state.KeyPair != nil {
		boshDeployInput.EC2KeyPair.Name = state.KeyPair.Name
		boshDeployInput.EC2KeyPair.PrivateKey = state.KeyPair.PrivateKey
		boshDeployInput.EC2KeyPair.PublicKey = state.KeyPair.PublicKey
	}

	if state.BOSH != nil {
		boshDeployInput.DirectorUsername = state.BOSH.DirectorUsername
		boshDeployInput.DirectorPassword = state.BOSH.DirectorPassword
		boshDeployInput.State = state.BOSH.State
		boshDeployInput.Credentials = state.BOSH.Credentials
		boshDeployInput.SSLKeyPair.Certificate = []byte(state.BOSH.DirectorSSLCertificate)
		boshDeployInput.SSLKeyPair.PrivateKey = []byte(state.BOSH.DirectorSSLPrivateKey)
	}

	if boshDeployInput.DirectorUsername == "" {
		if boshDeployInput.DirectorUsername, err = stringGenerator.Generate(USERNAME_PREFIX, USERNAME_LENGTH); err != nil {
			return BOSHDeployInput{}, err
		}
	}

	if boshDeployInput.DirectorPassword == "" {
		if boshDeployInput.DirectorPassword, err = stringGenerator.Generate(PASSWORD_PREFIX, PASSWORD_LENGTH); err != nil {
			return BOSHDeployInput{}, err
		}
	}
	return boshDeployInput, nil
}
