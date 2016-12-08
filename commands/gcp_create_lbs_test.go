package commands_test

import (
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCPCreateLBs", func() {
	Describe("Execute", func() {
		It("no-ops", func() {
			command := commands.NewGCPCreateLBs()
			err := command.Execute([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
