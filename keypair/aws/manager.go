package aws

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/keypair"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Manager struct {
	keyPairSynchronizer keyPairSynchronizer
	keyPairDeleter      keyPairDeleter
	clientProvider      clientProvider
}

type keyPairSynchronizer interface {
	Sync(ec2.KeyPair) (ec2.KeyPair, error)
}

type keyPairDeleter interface {
	Delete(keyPairName string) error
}

type clientProvider interface {
	SetConfig(config aws.Config)
}

func NewManager(keyPairSynchronizer keyPairSynchronizer, keyPairDeleter keyPairDeleter, clientProvider clientProvider) Manager {
	return Manager{
		keyPairSynchronizer: keyPairSynchronizer,
		keyPairDeleter:      keyPairDeleter,
		clientProvider:      clientProvider,
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

func (m Manager) Rotate(state storage.State) (storage.State, error) {
	if state.KeyPair.IsEmpty() {
		return storage.State{}, errors.New("no key found to rotate")
	}

	m.clientProvider.SetConfig(aws.Config{
		AccessKeyID:     state.AWS.AccessKeyID,
		SecretAccessKey: state.AWS.SecretAccessKey,
		Region:          state.AWS.Region,
	})

	err := m.keyPairDeleter.Delete(state.KeyPair.Name)
	if err != nil {
		return storage.State{}, err
	}

	keyPair, err := m.keyPairSynchronizer.Sync(ec2.KeyPair{
		Name: state.KeyPair.Name,
	})
	if err != nil {
		return storage.State{}, err
	}
	state.KeyPair.PrivateKey = keyPair.PrivateKey
	state.KeyPair.PublicKey = keyPair.PublicKey

	return state, nil
}
