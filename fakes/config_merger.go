package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/config"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Merger struct {
	MergeCall struct {
		CallCount int
		Receives  struct {
			GlobalFlags config.GlobalFlags
			State       storage.State
		}
		Returns struct {
			State storage.State
			Error error
		}
	}
}

func (f *Merger) MergeGlobalFlagsToState(globalFlags config.GlobalFlags, state storage.State) (storage.State, error) {
	f.MergeCall.CallCount++
	f.MergeCall.Receives.GlobalFlags = globalFlags
	f.MergeCall.Receives.State = state

	return f.MergeCall.Returns.State, f.MergeCall.Returns.Error
}
