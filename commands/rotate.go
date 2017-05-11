package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

const (
	RotateCommand = "rotate"
)

type Rotate struct {
	stateStore     stateStore
	keyPairManager keyPairManager
}

func NewRotate(stateStore stateStore, keyPairManager keyPairManager) Rotate {
	return Rotate{
		stateStore:     stateStore,
		keyPairManager: keyPairManager,
	}
}

func (r Rotate) Execute(args []string, state storage.State) error {
	state, err := r.keyPairManager.Rotate(state)
	if err != nil {
		return err
	}

	err = r.stateStore.Set(state)
	if err != nil {
		return err
	}

	return nil
}
