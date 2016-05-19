package commands

import "github.com/pivotal-cf-experimental/bosh-bootloader/storage"

type Command interface {
	Execute(subcommandFlags []string, state storage.State) (storage.State, error)
}
