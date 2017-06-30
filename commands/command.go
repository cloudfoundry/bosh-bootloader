package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

type Command interface {
	CheckFastFails(subcommandFlags []string, state storage.State) error
	Execute(subcommandFlags []string, state storage.State) error
	Usage() string
}

func lbExists(lbType string) bool {
	return lbType == "concourse" || lbType == "cf"
}
