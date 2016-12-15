package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
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
		)

		BeforeEach(func() {
			gcpDeleteLBs = &fakes.GCPDeleteLBs{}
			awsDeleteLBs = &fakes.AWSDeleteLBs{}
			stateValidator = &fakes.StateValidator{}
			logger = &fakes.Logger{}

			command = commands.NewDeleteLBs(gcpDeleteLBs, awsDeleteLBs, logger, stateValidator)
		})

		Context("when iaas is gcp", func() {
			It("calls gcp delete lbs", func() {
				err := command.Execute([]string{}, storage.State{
					IAAS: "gcp",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(gcpDeleteLBs.ExecuteCall.CallCount).To(Equal(1))
				Expect(gcpDeleteLBs.ExecuteCall.Receives.State).To(Equal(storage.State{
					IAAS: "gcp",
				}))
				Expect(awsDeleteLBs.ExecuteCall.CallCount).To(Equal(0))
			})
		})

		Context("when iaas is aws", func() {
			It("calls aws delete lbs", func() {
				err := command.Execute([]string{}, storage.State{
					IAAS: "aws",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(awsDeleteLBs.ExecuteCall.CallCount).To(Equal(1))
				Expect(awsDeleteLBs.ExecuteCall.Receives.State).To(Equal(storage.State{
					IAAS: "aws",
				}))
				Expect(gcpDeleteLBs.ExecuteCall.CallCount).To(Equal(0))
			})
		})

		Context("when --skip-if-missing is provided", func() {
			It("no-ops when lb does not exist", func() {
				err := command.Execute([]string{
					"--skip-if-missing",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(awsDeleteLBs.ExecuteCall.CallCount).To(Equal(0))
				Expect(gcpDeleteLBs.ExecuteCall.CallCount).To(Equal(0))

				Expect(logger.PrintlnCall.Receives.Message).To(Equal(`no lb type exists, skipping...`))
			})
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
