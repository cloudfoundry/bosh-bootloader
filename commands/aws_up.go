package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AWSUp struct{}

func NewAWSUp() AWSUp {
	return AWSUp{}
}

func (u AWSUp) Execute(state storage.State) (storage.State, error) {
	return state, nil
}
