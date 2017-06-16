package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

const LatestErrorCommand = "latest-error"

type LatestError struct {
	logger logger
}

func NewLatestError(logger logger) LatestError {
	return LatestError{
		logger: logger,
	}
}

func (l LatestError) CheckFastFails(subcommandFlags []string, state storage.State) error {
	return nil
}

func (l LatestError) Execute(subcommandFlags []string, bblState storage.State) error {
	l.logger.Println(bblState.LatestTFOutput)
	return nil
}
