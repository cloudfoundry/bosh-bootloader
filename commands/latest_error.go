package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

const LatestErrorCommand = "latest-error"

type LatestError struct {
	logger         logger
	stateValidator stateValidator
}

func NewLatestError(logger logger, stateValidator stateValidator) LatestError {
	return LatestError{
		logger:         logger,
		stateValidator: stateValidator,
	}
}

func (l LatestError) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := l.stateValidator.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (l LatestError) Execute(subcommandFlags []string, bblState storage.State) error {
	l.logger.Println(bblState.LatestTFOutput)
	return nil
}
