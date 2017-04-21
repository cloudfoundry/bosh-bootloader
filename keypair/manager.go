package keypair

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Manager struct {
	awsManager keyPairManager
	gcpManager keyPairManager
}

type keyPairManager interface {
	Sync(state storage.State) (storage.State, error)
}

func NewManager(awsManager keyPairManager, gcpManager keyPairManager) Manager {
	return Manager{
		awsManager: awsManager,
		gcpManager: gcpManager,
	}
}

func (m Manager) Sync(state storage.State) (storage.State, error) {
	switch state.IAAS {
	case "aws":
		return m.awsManager.Sync(state)
	case "gcp":
		return m.gcpManager.Sync(state)
	default:
		return storage.State{}, fmt.Errorf("invalid iaas was provided: %s", state.IAAS)
	}
}
