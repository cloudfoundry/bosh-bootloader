package commands

import "github.com/pivotal-cf-experimental/bosh-bootloader/storage"

type Command interface {
	Execute(GlobalFlags, storage.State) (storage.State, error)
}
