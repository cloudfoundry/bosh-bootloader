package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

type GCPUp struct {
	stateStore stateStore
}

func NewGCPUp(stateStore stateStore) GCPUp {
	return GCPUp{
		stateStore: stateStore,
	}
}

func (u GCPUp) Execute(state storage.State) error {
	state.IAAS = "gcp"

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	return nil
}
