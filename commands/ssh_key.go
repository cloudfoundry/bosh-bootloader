package commands

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	yaml "gopkg.in/yaml.v2"
)

const (
	SSHKeyCommand = "ssh-key"
)

type variables struct {
	JumpboxSSH struct {
		PrivateKey string `yaml:"private_key"`
	} `yaml:"jumpbox_ssh"`
}

type SSHKey struct {
	logger         logger
	stateValidator stateValidator
}

var unmarshal = yaml.Unmarshal

func NewSSHKey(logger logger, stateValidator stateValidator) SSHKey {
	return SSHKey{
		logger:         logger,
		stateValidator: stateValidator,
	}
}

func (s SSHKey) Execute(subcommandFlags []string, state storage.State) error {
	err := s.stateValidator.Validate()
	if err != nil {
		return err
	}

	v := variables{}
	err = unmarshal([]byte(state.BOSH.Variables), &v)
	if err != nil {
		return err
	}

	if v.JumpboxSSH.PrivateKey == "" {
		return errors.New("Could not retrieve the ssh key, please make sure you are targeting the proper state dir.")
	}

	s.logger.Println(v.JumpboxSSH.PrivateKey)

	return nil
}
