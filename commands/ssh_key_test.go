package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHKey", func() {
	var (
		sshKeyCommand commands.SSHKey

		stateValidator *fakes.StateValidator
		logger         *fakes.Logger

		incomingState storage.State
	)

	BeforeEach(func() {
		stateValidator = &fakes.StateValidator{}
		logger = &fakes.Logger{}

		incomingState = storage.State{
			BOSH: storage.BOSH{
				Variables: "jumpbox_ssh:\n  private_key: some-private-ssh-key",
			},
		}

		sshKeyCommand = commands.NewSSHKey(logger, stateValidator)
	})

	Describe("Execute", func() {
		It("validates state", func() {
			err := sshKeyCommand.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
		})

		It("prints the private ssh key of the jumpbox user", func() {
			err := sshKeyCommand.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.PrintlnCall.Messages).To(Equal([]string{"some-private-ssh-key"}))
		})

		Context("failure cases", func() {
			It("returns an error when state validator fails", func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
				err := sshKeyCommand.Execute([]string{}, incomingState)
				Expect(err).To(MatchError("state validator failed"))
			})

			It("returns an error when yaml unmarshal fails", func() {
				commands.SetUnmarshal(func([]byte, interface{}) error {
					return errors.New("yaml unmarshal failed")
				})
				err := sshKeyCommand.Execute([]string{}, incomingState)
				Expect(err).To(MatchError("yaml unmarshal failed"))
				commands.ResetUnmarshal()
			})

			It("returns an error when the jumpbox ssh private key is empty", func() {
				incomingState = storage.State{
					BOSH: storage.BOSH{
						Variables: "some_other_key:\n value: pair",
					},
				}
				err := sshKeyCommand.Execute([]string{}, incomingState)
				Expect(err).To(MatchError("Could not retrieve the ssh key, please make sure you are targeting the proper state dir."))
			})
		})
	})
})
