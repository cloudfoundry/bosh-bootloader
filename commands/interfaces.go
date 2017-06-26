package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

type terraformApplier interface {
	ValidateVersion() error
	Import(initialState storage.State, stackOutputs map[string]string) (storage.State, error)
	Apply(storage.State) (storage.State, error)
}

type terraformDestroyer interface {
	ValidateVersion() error
	GetOutputs(storage.State) (map[string]interface{}, error)
	Destroy(storage.State) (storage.State, error)
}

type terraformOutputter interface {
	GetOutputs(storage.State) (map[string]interface{}, error)
}
