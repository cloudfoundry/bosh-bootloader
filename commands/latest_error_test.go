package commands_test

import (
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("latest-error", func() {
	var (
		command commands.LatestError
		logger  *fakes.Logger
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}

		command = commands.NewLatestError(logger)
	})

	Describe("CheckFastFails", func() {
		It("returns no error", func() {
			err := command.CheckFastFails([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Execute", func() {
		It("prints the latest terraform output", func() {
			bblState := storage.State{
				LatestTFOutput: "some tf output",
			}

			err := command.Execute([]string{}, bblState)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.PrintlnCall.Messages).To(ContainElement("some tf output"))
		})
	})
})
