package commands

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	yaml "gopkg.in/yaml.v2"
)

const (
	SSHKeyCommand = "ssh-key"
)

type SSHKey struct {
	logger              logger
	stateValidator      stateValidator
	jumpboxSSHKeyGetter jumpboxSSHKeyGetter
}

type jumpboxSSHKeyGetter interface {
	Get(storage.State) (string, error)
}

var unmarshal = yaml.Unmarshal

func NewSSHKey(logger logger, stateValidator stateValidator, jumpboxSSHKeyGetter jumpboxSSHKeyGetter) SSHKey {
	return SSHKey{
		logger:              logger,
		stateValidator:      stateValidator,
		jumpboxSSHKeyGetter: jumpboxSSHKeyGetter,
	}
}

func (s SSHKey) Execute(subcommandFlags []string, state storage.State) error {
	err := s.stateValidator.Validate()
	if err != nil {
		return err
	}

	privateKey, err := s.jumpboxSSHKeyGetter.Get(state)
	if err != nil {
		return err
	}

	if privateKey == "" {
		return errors.New("Could not retrieve the ssh key, please make sure you are targeting the proper state dir.")
	}

	s.logger.Println(privateKey)

	return nil
}
