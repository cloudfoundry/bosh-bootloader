package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("latest-error", func() {
	var (
		logger         *fakes.Logger
		stateValidator *fakes.StateValidator

		command commands.LatestError
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		stateValidator = &fakes.StateValidator{}

		command = commands.NewLatestError(logger, stateValidator)
	})

	Describe("CheckFastFails", func() {
		It("returns an error when the state does not exist", func() {
			stateValidator.ValidateCall.Returns.Error = errors.New("failed to validate state")
			err := command.CheckFastFails([]string{}, storage.State{})
			Expect(err).To(MatchError("failed to validate state"))
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
