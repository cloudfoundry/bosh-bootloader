package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteLBs", func() {
	var (
		command commands.DeleteLBs

		deleteLBs      *fakes.DeleteLBs
		stateValidator *fakes.StateValidator
		logger         *fakes.Logger
		boshManager    *fakes.BOSHManager
	)

	BeforeEach(func() {
		deleteLBs = &fakes.DeleteLBs{}
		stateValidator = &fakes.StateValidator{}
		logger = &fakes.Logger{}
		boshManager = &fakes.BOSHManager{}
		boshManager.VersionCall.Returns.Version = "2.0.24"

		command = commands.NewDeleteLBs(deleteLBs, logger, stateValidator, boshManager)
	})

	Describe("CheckFastFails", func() {
		Context("when state validator fails", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			})

			It("returns an error", func() {
				err := command.CheckFastFails([]string{}, storage.State{})

				Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(err).To(MatchError("state validator failed"))
			})
		})

		Context("when the BOSH version is less than 2.0.24 and there is a director", func() {
			It("returns a helpful error message", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := command.CheckFastFails([]string{}, storage.State{
					IAAS: "aws",
					LB: storage.LB{
						Type: "concourse",
					},
				})
				Expect(err).To(MatchError("BOSH version must be at least v2.0.24"))
			})
		})

		Context("when the BOSH version is less than 2.0.24 and there is no director", func() {
			It("does not fast fail", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := command.CheckFastFails([]string{}, storage.State{
					IAAS:       "gcp",
					NoDirector: true,
					LB: storage.LB{
						Type: "concourse",
					},
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Execute", func() {
		It("calls  delete lbs", func() {
			err := command.Execute([]string{}, storage.State{
				LB: storage.LB{
					Type: "concourse",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(deleteLBs.ExecuteCall.CallCount).To(Equal(1))
			Expect(deleteLBs.ExecuteCall.Receives.State).To(Equal(storage.State{
				LB: storage.LB{
					Type: "concourse",
				},
			}))
		})

		Context("when --skip-if-missing is provided", func() {
			DescribeTable("no-ops", func(state storage.State) {
				err := command.Execute([]string{
					"--skip-if-missing",
				}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(deleteLBs.ExecuteCall.CallCount).To(Equal(0))
				Expect(logger.PrintlnCall.Receives.Message).To(Equal(`no lb type exists, skipping...`))
			},
				Entry("no-ops when LB type does not exist in state LB", storage.State{
					LB: storage.LB{
						Type: "",
					},
				}),
			)
		})

		Context("failure cases", func() {
			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--unknown-flag"}, storage.State{})
					Expect(err).To(MatchError("flag provided but not defined: -unknown-flag"))

					Expect(deleteLBs.ExecuteCall.CallCount).To(Equal(0))
				})
			})
		})
	})
})
