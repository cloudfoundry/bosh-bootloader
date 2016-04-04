package commands

import "github.com/pivotal-cf-experimental/bosh-bootloader/storage"

type Command interface {
	Execute(GlobalFlags, []string, storage.State) (storage.State, error)
}
