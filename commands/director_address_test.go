package commands_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("director-address", func() {
	var (
		command    commands.DirectorAddress
		fakeLogger *fakes.Logger
	)
	BeforeEach(func() {
		fakeLogger = &fakes.Logger{}
		command = commands.NewDirectorAddress(fakeLogger)
	})

	Describe("Execute", func() {
		It("prints out the director address", func() {
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

			state, err := command.Execute(commands.GlobalFlags{}, []string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(state).To(Equal(incomingState))
		})

		Context("failure cases", func() {
			Context("returns an error when the director address does not exist", func() {
				It("does not panic when the state is empty", func() {
					_, err := command.Execute(commands.GlobalFlags{}, []string{}, storage.State{})
					Expect(err).To(MatchError("Could not retrieve director address, please make sure you are targeting the proper state dir."))

					Expect(fakeLogger.PrintlnCall.CallCount).To(Equal(0))
				})

				It("returns a helpful error message", func() {
					_, err := command.Execute(commands.GlobalFlags{}, []string{}, storage.State{
						BOSH: storage.BOSH{},
					})
					Expect(err).To(MatchError("Could not retrieve director address, please make sure you are targeting the proper state dir."))

					Expect(fakeLogger.PrintlnCall.CallCount).To(Equal(0))
				})
			})
		})
	})
})
