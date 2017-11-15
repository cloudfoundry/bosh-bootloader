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

		stateValidator *fakes.StateValidator
		logger         *fakes.Logger
		sshKeyGetter   *fakes.SSHKeyGetter
	)

	BeforeEach(func() {
		incomingState = storage.State{}

		stateValidator = &fakes.StateValidator{}
		logger = &fakes.Logger{}
		sshKeyGetter = &fakes.SSHKeyGetter{}
		sshKeyGetter.GetCall.Returns.PrivateKey = "some-private-ssh-key"

		sshKeyCommand = commands.NewSSHKey(logger, stateValidator, sshKeyGetter)
	})

	Describe("CheckFastFails", func() {
		Context("when state validator fails", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			})

			It("returns an error", func() {
				err := sshKeyCommand.CheckFastFails([]string{}, incomingState)
				Expect(err).To(MatchError("state validator failed"))
			})
		})
	})

	Describe("Execute", func() {
		It("prints the private ssh key of the jumpbox user", func() {
			err := sshKeyCommand.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(sshKeyGetter.GetCall.CallCount).To(Equal(1))
			Expect(sshKeyGetter.GetCall.Receives.Deployment).To(Equal("jumpbox"))
			Expect(logger.PrintlnCall.Messages).To(Equal([]string{"some-private-ssh-key"}))
		})

		Context("director-ssh-key", func() {
			BeforeEach(func() {
				sshKeyCommand = commands.NewDirectorSSHKey(logger, stateValidator, sshKeyGetter)
			})

			It("uses BOSH variables to get the SSH key", func() {
				err := sshKeyCommand.Execute([]string{}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(sshKeyGetter.GetCall.CallCount).To(Equal(1))
				Expect(sshKeyGetter.GetCall.Receives.Deployment).To(Equal("director"))
			})
		})

		Context("failure cases", func() {
			Context("when the ssh key getter fails", func() {
				BeforeEach(func() {
					sshKeyGetter.GetCall.Returns.Error = errors.New("jumpbox ssh key getter failed")
				})

				It("returns an error", func() {
					err := sshKeyCommand.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("jumpbox ssh key getter failed"))
				})
			})

			Context("when the ssh private key is empty", func() {
				BeforeEach(func() {
					sshKeyGetter.GetCall.Returns.PrivateKey = ""
				})

				It("returns an error", func() {
					err := sshKeyCommand.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("Could not retrieve the ssh key, please make sure you are targeting the proper state dir."))
				})
			})
		})
	})
})
