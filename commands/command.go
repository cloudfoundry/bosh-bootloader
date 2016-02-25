package commands

import "github.com/pivotal-cf-experimental/bosh-bootloader/state"

type Command interface {
	Execute(GlobalFlags, state.State) (state.State, error)
}
