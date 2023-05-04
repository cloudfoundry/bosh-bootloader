package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LBs", func() {
	var (
		lbsCommand commands.LBs

		lbs            *fakes.LBs
		stateValidator *fakes.StateValidator
	)

	BeforeEach(func() {
		lbs = &fakes.LBs{}
		stateValidator = &fakes.StateValidator{}

		lbsCommand = commands.NewLBs(lbs, stateValidator)
	})

	Describe("CheckFastFails", func() {
		Context("when state validator fails", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			})

			It("returns an error", func() {
				err := lbsCommand.CheckFastFails([]string{}, storage.State{})

				Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(err).To(MatchError("state validator failed"))
			})
		})
	})

	Describe("Execute", func() {
		It("prints LB ips", func() {
			incomingState := storage.State{
				IAAS: "aws",
			}
			err := lbsCommand.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(lbs.ExecuteCall.Receives.SubcommandFlags).To(Equal([]string{}))
			Expect(lbs.ExecuteCall.Receives.State).To(Equal(incomingState))
		})

		Context("failure cases", func() {
			Context("when LBs fails", func() {
				BeforeEach(func() {
					lbs.ExecuteCall.Returns.Error = errors.New("something bad happened")
				})

				It("returns an error", func() {
					err := lbsCommand.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("something bad happened"))
				})
			})
		})
	})
})
