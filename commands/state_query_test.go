package commands_test

import (
	"fmt"
	"math/rand"

	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StateQuery", func() {
	var (
		fakeLogger *fakes.Logger
	)

	BeforeEach(func() {
		fakeLogger = &fakes.Logger{}
	})

	Describe("Execute", func() {
		It("prints out the director address", func() {
			command := commands.NewStateQuery(fakeLogger, "director address", func(state storage.State) string {
				return state.BOSH.DirectorAddress
			})

			state := storage.State{
				BOSH: storage.BOSH{
					DirectorAddress: "some-director-address",
				},
			}

			_, err := command.Execute(commands.GlobalFlags{}, []string{}, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLogger.PrintlnCall.Receives.Message).To(Equal("some-director-address"))
		})

		It("returns the given state unmodified", func() {
			incomingState := storage.State{
				BOSH: storage.BOSH{
					DirectorAddress: "some-director-address",
				},
			}

			command := commands.NewStateQuery(fakeLogger, "director address", func(state storage.State) string {
				return incomingState.BOSH.DirectorAddress
			})

			state, err := command.Execute(commands.GlobalFlags{}, []string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(state).To(Equal(incomingState))
		})

		It("returns an error when the state value is empty", func() {
			propertyName := fmt.Sprintf("%s-%d", "some-name", rand.Int())
			command := commands.NewStateQuery(fakeLogger, propertyName, func(state storage.State) string {
				return ""
			})
			_, err := command.Execute(commands.GlobalFlags{}, []string{}, storage.State{
				BOSH: storage.BOSH{},
			})
			Expect(err).To(MatchError(fmt.Sprintf("Could not retrieve %s, please make sure you are targeting the proper state dir.", propertyName)))

			Expect(fakeLogger.PrintlnCall.CallCount).To(Equal(0))
		})
	})
})
