package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

type GCPCreateLBs struct{}

func NewGCPCreateLBs() GCPCreateLBs {
	return GCPCreateLBs{}
}

func (GCPCreateLBs) Execute([]string, storage.State) error {
	return nil
}
