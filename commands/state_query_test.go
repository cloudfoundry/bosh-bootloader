package commands_test

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StateQuery", func() {
	var (
		fakeLogger         *fakes.Logger
		fakeStateValidator *fakes.StateValidator
	)

	BeforeEach(func() {
		fakeLogger = &fakes.Logger{}
		fakeStateValidator = &fakes.StateValidator{}
	})

	Describe("Execute", func() {
		It("prints out the director address", func() {
			command := commands.NewStateQuery(fakeLogger, fakeStateValidator, "director address", func(state storage.State) string {
				return state.BOSH.DirectorAddress
			})

			state := storage.State{
				BOSH: storage.BOSH{
					DirectorAddress: "some-director-address",
				},
			}

			err := command.Execute([]string{}, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLogger.PrintlnCall.Receives.Message).To(Equal("some-director-address"))
		})

		It("returns an error when the state validator fails", func() {
			fakeStateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			command := commands.NewStateQuery(fakeLogger, fakeStateValidator, "", func(state storage.State) string {
				return ""
			})

			err := command.Execute([]string{}, storage.State{
				BOSH: storage.BOSH{},
			})

			Expect(err).To(MatchError("state validator failed"))
		})

		It("returns an error when the state value is empty", func() {
			propertyName := fmt.Sprintf("%s-%d", "some-name", rand.Int())
			command := commands.NewStateQuery(fakeLogger, fakeStateValidator, propertyName, func(state storage.State) string {
				return ""
			})
			err := command.Execute([]string{}, storage.State{
				BOSH: storage.BOSH{},
			})
			Expect(err).To(MatchError(fmt.Sprintf("Could not retrieve %s, please make sure you are targeting the proper state dir.", propertyName)))

			Expect(fakeLogger.PrintlnCall.CallCount).To(Equal(0))
		})
	})
})
