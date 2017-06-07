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

		incomingState storage.State

		stateValidator      *fakes.StateValidator
		logger              *fakes.Logger
		jumpboxSSHKeyGetter *fakes.JumpboxSSHKeyGetter
	)

	BeforeEach(func() {
		incomingState = storage.State{
			Version: 3,
		}

		stateValidator = &fakes.StateValidator{}
		logger = &fakes.Logger{}
		jumpboxSSHKeyGetter = &fakes.JumpboxSSHKeyGetter{}
		jumpboxSSHKeyGetter.GetCall.Returns.PrivateKey = "some-private-ssh-key"

		sshKeyCommand = commands.NewSSHKey(logger, stateValidator, jumpboxSSHKeyGetter)
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

			Expect(jumpboxSSHKeyGetter.GetCall.CallCount).To(Equal(1))
			Expect(jumpboxSSHKeyGetter.GetCall.Receives.State).To(Equal(incomingState))
			Expect(logger.PrintlnCall.Messages).To(Equal([]string{"some-private-ssh-key"}))
		})

		Context("failure cases", func() {
			It("returns an error when state validator fails", func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
				err := sshKeyCommand.Execute([]string{}, incomingState)
				Expect(err).To(MatchError("state validator failed"))
			})

			It("returns an error when the jumpbox ssh key getter fails", func() {
				jumpboxSSHKeyGetter.GetCall.Returns.Error = errors.New("jumpbox ssh key getter failed")
				err := sshKeyCommand.Execute([]string{}, incomingState)
				Expect(err).To(MatchError("jumpbox ssh key getter failed"))
			})

			It("returns an error when the jumpbox ssh private key is empty", func() {
				jumpboxSSHKeyGetter.GetCall.Returns.PrivateKey = ""
				err := sshKeyCommand.Execute([]string{}, incomingState)
				Expect(err).To(MatchError("Could not retrieve the ssh key, please make sure you are targeting the proper state dir."))
			})
		})
	})
})
