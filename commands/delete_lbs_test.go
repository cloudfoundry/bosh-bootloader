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
	Describe("Execute", func() {
		var (
			command commands.DeleteLBs

			gcpDeleteLBs   *fakes.GCPDeleteLBs
			awsDeleteLBs   *fakes.AWSDeleteLBs
			stateValidator *fakes.StateValidator
			logger         *fakes.Logger
			boshManager    *fakes.BOSHManager
		)

		BeforeEach(func() {
			gcpDeleteLBs = &fakes.GCPDeleteLBs{}
			awsDeleteLBs = &fakes.AWSDeleteLBs{}
			stateValidator = &fakes.StateValidator{}
			logger = &fakes.Logger{}
			boshManager = &fakes.BOSHManager{}
			boshManager.VersionCall.Returns.Version = "2.0.0"

			command = commands.NewDeleteLBs(gcpDeleteLBs, awsDeleteLBs, logger, stateValidator, boshManager)
		})

		Context("when the BOSH version is less than 2.0.0 and there is a director", func() {
			It("returns a helpful error message", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := command.Execute([]string{}, storage.State{
					IAAS: "aws",
					LB: storage.LB{
						Type: "concourse",
					},
				})
				Expect(err).To(MatchError("BOSH version must be at least v2.0.0"))
			})
		})

		Context("when the BOSH version is less than 2.0.0 and there is no director", func() {
			It("does not fast fail", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := command.Execute([]string{}, storage.State{
					IAAS:       "gcp",
					NoDirector: true,
					LB: storage.LB{
						Type: "concourse",
					},
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when iaas is gcp", func() {
			It("calls gcp delete lbs", func() {
				err := command.Execute([]string{}, storage.State{
					IAAS: "gcp",
					LB: storage.LB{
						Type: "concourse",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(gcpDeleteLBs.ExecuteCall.CallCount).To(Equal(1))
				Expect(gcpDeleteLBs.ExecuteCall.Receives.State).To(Equal(storage.State{
					IAAS: "gcp",
					LB: storage.LB{
						Type: "concourse",
					},
				}))
				Expect(awsDeleteLBs.ExecuteCall.CallCount).To(Equal(0))
			})
		})

		Context("when iaas is aws", func() {
			It("calls aws delete lbs", func() {
				err := command.Execute([]string{}, storage.State{
					IAAS: "aws",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(awsDeleteLBs.ExecuteCall.CallCount).To(Equal(1))
				Expect(awsDeleteLBs.ExecuteCall.Receives.State).To(Equal(storage.State{
					IAAS: "aws",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				}))
				Expect(gcpDeleteLBs.ExecuteCall.CallCount).To(Equal(0))
			})
		})

		Context("when --skip-if-missing is provided", func() {
			DescribeTable("no-ops", func(state storage.State) {
				err := command.Execute([]string{
					"--skip-if-missing",
				}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(awsDeleteLBs.ExecuteCall.CallCount).To(Equal(0))
				Expect(gcpDeleteLBs.ExecuteCall.CallCount).To(Equal(0))

				Expect(logger.PrintlnCall.Receives.Message).To(Equal(`no lb type exists, skipping...`))
			},
				Entry("no-ops when LB type does not exist in state stack", storage.State{
					Stack: storage.Stack{
						LBType: "",
					},
				}),
				Entry("no-ops when LB type does not exist in state LB", storage.State{
					LB: storage.LB{
						Type: "",
					},
				}),
			)

			DescribeTable("deletes the LB", func(state storage.State) {
				err := command.Execute([]string{
					"--skip-if-missing",
				}, state)
				Expect(err).NotTo(HaveOccurred())

				if state.IAAS == "aws" {
					Expect(awsDeleteLBs.ExecuteCall.CallCount).To(Equal(1))
				} else {
					Expect(gcpDeleteLBs.ExecuteCall.CallCount).To(Equal(1))
				}
			},
				Entry("deletes the LB when LB type exists in state stack", storage.State{
					IAAS: "aws",
					Stack: storage.Stack{
						LBType: "concourse",
					},
				}),
				Entry("deletes the LB when LB type exists in state LB", storage.State{
					IAAS: "gcp",
					LB: storage.LB{
						Type: "concourse",
					},
				}),
			)
		})

		Context("failure cases", func() {
			It("returns an error when an unknown flag is provided", func() {
				err := command.Execute([]string{"--unknown-flag"}, storage.State{})
				Expect(err).To(MatchError("flag provided but not defined: -unknown-flag"))

				Expect(awsDeleteLBs.ExecuteCall.CallCount).To(Equal(0))
				Expect(gcpDeleteLBs.ExecuteCall.CallCount).To(Equal(0))
			})

			It("returns an error when an unknown iaas is in the state", func() {
				err := command.Execute([]string{}, storage.State{
					IAAS: "some-unknown-iaas",
				})
				Expect(err).To(MatchError(`"some-unknown-iaas" is an invalid iaas type in state, supported iaas types are: [gcp, aws]`))
			})

			It("returns an error when state validator fails", func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
				err := command.Execute([]string{}, storage.State{})

				Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(err).To(MatchError("state validator failed"))
			})
		})
	})
})
