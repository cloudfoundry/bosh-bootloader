package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type FilteredDeleter interface {
	Delete(filter string) error
}

type CleanupLeftovers struct {
	deleter FilteredDeleter
}

func NewCleanupLeftovers(deleter FilteredDeleter) CleanupLeftovers {
	return CleanupLeftovers{
		deleter: deleter,
	}
}

func (l CleanupLeftovers) CheckFastFails(subcommandFlags []string, state storage.State) error {
	return nil
}

func (l CleanupLeftovers) Execute(subcommandFlags []string, state storage.State) error {
	var filter string
	f := flags.New("cleanup-leftovers")
	f.String(&filter, "filter", "")

	err := f.Parse(subcommandFlags)
	if err != nil {
		return fmt.Errorf("Parsing cleanup-leftovers args: %s", err)
	}

	return l.deleter.Delete(filter)
}

func (l CleanupLeftovers) Usage() string {
	return fmt.Sprintf("%s%s%s", CleanupLeftoversCommandUsage, requiresCredentials, Credentials)
}
