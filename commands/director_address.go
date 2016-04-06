package commands

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type DirectorAddress struct {
	logger logger
}

func NewDirectorAddress(logger logger) DirectorAddress {
	return DirectorAddress{
		logger: logger,
	}
}

func (d DirectorAddress) Execute(globalFlags GlobalFlags, subcommandFlags []string, state storage.State) (storage.State, error) {
	if state.BOSH == nil || state.BOSH.DirectorAddress == "" {
		return state, errors.New("Could not retrieve director address, please make sure you are targeting the proper state dir.")
	}

	d.logger.Println(state.BOSH.DirectorAddress)
	return state, nil
}
