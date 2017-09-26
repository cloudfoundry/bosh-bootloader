package commands_test

import (
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWSUp", func() {
	Describe("Execute", func() {
		var (
			command       commands.AWSUp
			incomingState storage.State
		)

		BeforeEach(func() {
			incomingState = storage.State{IAAS: "aws"}
			command = commands.NewAWSUp()
		})

		It("returns the state it was called with", func() {
			returnedState, err := command.Execute(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(returnedState).To(Equal(incomingState))
		})
	})
})
