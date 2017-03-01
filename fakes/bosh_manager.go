package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type BOSHManager struct {
	CreateCall struct {
		CallCount int
		Receives  struct {
			State   storage.State
			OpsFile []byte
		}
		Returns struct {
			State storage.State
			Error error
		}
	}
	DeleteCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Error error
		}
	}
	GetDeploymentVarsCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Vars  string
			Error error
		}
	}
}

func (b *BOSHManager) Create(state storage.State, opsFile []byte) (storage.State, error) {
	b.CreateCall.CallCount++
	b.CreateCall.Receives.State = state
	b.CreateCall.Receives.OpsFile = opsFile
	state.BOSH = b.CreateCall.Returns.State.BOSH
	return state, b.CreateCall.Returns.Error
}

func (b *BOSHManager) Delete(state storage.State) error {
	b.DeleteCall.CallCount++
	b.DeleteCall.Receives.State = state
	return b.DeleteCall.Returns.Error
}

func (b *BOSHManager) GetDeploymentVars(state storage.State) (string, error) {
	b.GetDeploymentVarsCall.CallCount++
	b.GetDeploymentVarsCall.Receives.State = state
	return b.GetDeploymentVarsCall.Returns.Vars, b.GetDeploymentVarsCall.Returns.Error
}
