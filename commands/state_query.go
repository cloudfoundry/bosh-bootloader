package commands

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	EnvIDCommand            = "env-id"
	SSHKeyCommand           = "ssh-key"
	DirectorUsernameCommand = "director-username"
	DirectorPasswordCommand = "director-password"
	DirectorAddressCommand  = "director-address"
	DirectorCACertCommand   = "director-ca-cert"

	EnvIDPropertyName            = "environment id"
	SSHKeyPropertyName           = "ssh key"
	DirectorUsernamePropertyName = "director username"
	DirectorPasswordPropertyName = "director password"
	DirectorAddressPropertyName  = "director address"
	DirectorCACertPropertyName   = "director ca cert"
)

type StateQuery struct {
	logger         logger
	stateValidator stateValidator
	propertyName   string
}

type getPropertyFunc func(storage.State) string

func NewStateQuery(logger logger, stateValidator stateValidator, propertyName string) StateQuery {
	return StateQuery{
		logger:         logger,
		stateValidator: stateValidator,
		propertyName:   propertyName,
	}
}

func (s StateQuery) Execute(subcommandFlags []string, state storage.State) error {
	err := s.stateValidator.Validate()
	if err != nil {
		return err
	}

	if state.NoDirector {
		return errors.New("Error BBL does not manage this director.")
	}

	var propertyValue string
	switch s.propertyName {
	case DirectorAddressPropertyName:
		propertyValue = state.BOSH.DirectorAddress
	case DirectorUsernamePropertyName:
		propertyValue = state.BOSH.DirectorUsername
	case DirectorPasswordPropertyName:
		propertyValue = state.BOSH.DirectorPassword
	case DirectorCACertPropertyName:
		propertyValue = state.BOSH.DirectorSSLCA
	case SSHKeyPropertyName:
		propertyValue = state.KeyPair.PrivateKey
	case EnvIDPropertyName:
		propertyValue = state.EnvID
	}

	if propertyValue == "" {
		return fmt.Errorf("Could not retrieve %s, please make sure you are targeting the proper state dir.", s.propertyName)
	}

	s.logger.Println(propertyValue)
	return nil
}
