package commands

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type StateQuery struct {
	logger       logger
	propertyName string
	getProperty  getPropertyFunc
}

type getPropertyFunc func(storage.State) string

func NewStateQuery(logger logger, propertyName string, getProperty getPropertyFunc) StateQuery {
	return StateQuery{
		logger:       logger,
		propertyName: propertyName,
		getProperty:  getProperty,
	}
}

func (s StateQuery) Execute(subcommandFlags []string, state storage.State) error {
	propertyValue := s.getProperty(state)
	if propertyValue == "" {
		return fmt.Errorf("Could not retrieve %s, please make sure you are targeting the proper state dir.", s.propertyName)
	}

	s.logger.Println(propertyValue)
	return nil
}
