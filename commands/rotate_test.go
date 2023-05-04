package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rotate", func() {
	var (
		stateValidator *fakes.StateValidator
		sshKeyDeleter  *fakes.SSHKeyDeleter
		up             *fakes.Up
		rotate         commands.Rotate
	)

	BeforeEach(func() {
		stateValidator = &fakes.StateValidator{}
		sshKeyDeleter = &fakes.SSHKeyDeleter{}
		up = &fakes.Up{}
		rotate = commands.NewRotate(stateValidator, sshKeyDeleter, up)
	})

	Describe("CheckFastFails", func() {
		It("calls stateValidator.Validate", func() {
			err := rotate.CheckFastFails([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
		})

		It("calls up.CheckFastFails", func() {
			subcommandFlags := []string{"some", "subcommand", "flags"}
			state := storage.State{EnvID: "some-env-id"}
			err := rotate.CheckFastFails(subcommandFlags, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(up.CheckFastFailsCall.CallCount).To(Equal(1))
			Expect(up.CheckFastFailsCall.Receives.SubcommandFlags).To(Equal(subcommandFlags))
			Expect(up.CheckFastFailsCall.Receives.State).To(Equal(state))
		})

		Context("when the state validator returns an error", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("coconut")
			})

			It("calls stateValidator.Validate", func() {
				err := rotate.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("validate state: coconut"))
			})
		})

		Context("when up.CheckFastFails returns and error", func() {
			BeforeEach(func() {
				up.CheckFastFailsCall.Returns.Error = errors.New("passionfruit")
			})

			It("wraps and returns the error", func() {
				err := rotate.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("up: passionfruit"))
			})
		})

	})

	Describe("Execute", func() {
		var (
			state storage.State
			args  []string
		)

		BeforeEach(func() {
			args = []string{"some", "args"}
			state = storage.State{
				EnvID: "some-env-id",
				Jumpbox: storage.Jumpbox{
					Variables: "ssh_key: foo\nsome_other_key: bar\n",
				},
			}
		})

		It("deletes the ssh key", func() {
			err := rotate.Execute(args, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(sshKeyDeleter.DeleteCall.CallCount).To(Equal(1))
		})

		It("calls up with args and new state", func() {
			err := rotate.Execute(args, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(up.ExecuteCall.CallCount).To(Equal(1))
			Expect(up.ExecuteCall.Receives.Args).To(Equal(args))
			Expect(up.ExecuteCall.Receives.State).To(Equal(state))
		})

		Context("when the ssh key deleter returns an error", func() {
			BeforeEach(func() {
				sshKeyDeleter.DeleteCall.Returns.Error = errors.New("guava")
			})

			It("wraps and returns the error from the sshKeyDeleter", func() {
				err := rotate.Execute(args, state)
				Expect(err).To(MatchError("delete ssh key: guava"))
			})
		})

		Context("when up returns an error", func() {
			BeforeEach(func() {
				up.ExecuteCall.Returns.Error = errors.New("fig")
			})

			It("returns the error from up", func() {
				err := rotate.Execute(args, state)
				Expect(err).To(MatchError("up: fig"))
			})
		})
	})
})
