package commands

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	yaml "gopkg.in/yaml.v2"
)

type SSHKey struct {
	logger         logger
	stateValidator stateValidator
	sshKeyGetter   sshKeyGetter
	Director       bool
}

type sshKeyGetter interface {
	Get(string) (string, error)
}

var unmarshal = yaml.Unmarshal

func NewSSHKey(logger logger, stateValidator stateValidator, sshKeyGetter sshKeyGetter) SSHKey {
	return SSHKey{
		logger:         logger,
		stateValidator: stateValidator,
		sshKeyGetter:   sshKeyGetter,
	}
}

func NewDirectorSSHKey(logger logger, stateValidator stateValidator, sshKeyGetter sshKeyGetter) SSHKey {
	return SSHKey{
		logger:         logger,
		stateValidator: stateValidator,
		sshKeyGetter:   sshKeyGetter,
		Director:       true,
	}
}

func (s SSHKey) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := s.stateValidator.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (s SSHKey) Execute(subcommandFlags []string, state storage.State) error {
	deployment := "jumpbox"
	if s.Director {
		deployment = "director"
	}

	privateKey, err := s.sshKeyGetter.Get(deployment)
	if err != nil {
		return err
	}

	if privateKey == "" {
		return errors.New("Could not retrieve the ssh key, please make sure you are targeting the proper state dir.")
	}

	s.logger.Println(privateKey)

	return nil
}
