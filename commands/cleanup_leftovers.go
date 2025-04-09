package commands

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type FilteredDeleter interface {
	Delete(filter string) error
	List(filter string)
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
	var (
		filter string
		dryRun bool
	)
	f := flags.New("cleanup-leftovers")
	f.String(&filter, "filter", "")
	f.Bool(&dryRun, "dry-run")

	err := f.Parse(subcommandFlags)
	if err != nil {
		return fmt.Errorf("Parsing cleanup-leftovers args: %s", err) //nolint:staticcheck
	}

	if state.IAAS == "vsphere" && filter == "" {
		// vSphere requires a filter
		return errors.New("cleanup-leftovers on vSphere requires a filter.\nProvide a filter using the --filter or -f flag.") //nolint:staticcheck
	}

	if state.IAAS == "openstack" {
		// we don't create network infrastructure on openstack
		// and we don't tear it down either
		return nil
	}

	if dryRun {
		l.deleter.List(filter)
		return nil
	}

	return l.deleter.Delete(filter)
}

func (l CleanupLeftovers) Usage() string {
	return fmt.Sprintf("%s%s%s", CleanupLeftoversCommandUsage, requiresCredentials, Credentials)
}
