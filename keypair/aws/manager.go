package aws

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/keypair"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Manager struct {
	keyPairSynchronizer keyPairSynchronizer
}

type keyPairSynchronizer interface {
	Sync(ec2.KeyPair) (ec2.KeyPair, error)
}

func NewManager(keyPairSynchronizer keyPairSynchronizer) Manager {
	return Manager{
		keyPairSynchronizer: keyPairSynchronizer,
	}
}

func (m Manager) Sync(state storage.State) (storage.State, error) {
	if state.EnvID == "" {
		return storage.State{}, errors.New("env id must be set to generate a keypair")
	}

	if state.KeyPair.Name == "" {
		state.KeyPair.Name = fmt.Sprintf("keypair-%s", state.EnvID)
	}

	keyPair, err := m.keyPairSynchronizer.Sync(ec2.KeyPair{
		Name:       state.KeyPair.Name,
		PublicKey:  state.KeyPair.PublicKey,
		PrivateKey: state.KeyPair.PrivateKey,
	})
	if err != nil {
		return storage.State{}, keypair.NewManagerError(state, err)
	}

	state.KeyPair.PrivateKey = keyPair.PrivateKey
	state.KeyPair.PublicKey = keyPair.PublicKey

	return state, nil
}
