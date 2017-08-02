package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

type Rotate struct {
	stateStore     stateStore
	keyPairManager keyPairManager
	terraform      terraformOutputter
	boshManager    boshManager
	stateValidator stateValidator
}

func NewRotate(stateStore stateStore, keyPairManager keyPairManager, terraform terraformOutputter, boshManager boshManager, stateValidator stateValidator) Rotate {
	return Rotate{
		stateStore:     stateStore,
		keyPairManager: keyPairManager,
		terraform:      terraform,
		boshManager:    boshManager,
		stateValidator: stateValidator,
	}
}

func (r Rotate) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := r.stateValidator.Validate()
	if err != nil {
		return err
	}

	return nil
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

	terraformOutputs, err := r.terraform.GetOutputs(state)
	if err != nil {
		return err
	}

	if !state.NoDirector {
		state, err = r.boshManager.CreateDirector(state, terraformOutputs)
		if err != nil {
			return err
		}

		err = r.stateStore.Set(state)
		if err != nil {
			return err
		}
	}

	return nil
}
